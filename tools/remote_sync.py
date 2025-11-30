#!/usr/bin/env python3
"""Simple incremental sync client for pushing honeypot data to the central API."""

import argparse
import json
import logging
import os
import sys
import time
from datetime import date, datetime
from decimal import Decimal
from typing import Any, Dict, List, Tuple

import pymysql
import requests

LOGGER = logging.getLogger("remote-sync")


def load_config(path: str) -> Dict[str, Any]:
    with open(path, "r", encoding="utf-8") as handler:
        return json.load(handler)


def load_state(path: str) -> Dict[str, int]:
    if not os.path.exists(path):
        return {}
    with open(path, "r", encoding="utf-8") as handler:
        try:
            return json.load(handler)
        except json.JSONDecodeError:
            return {}


def save_state(path: str, state: Dict[str, int]) -> None:
    directory = os.path.dirname(path)
    if directory and not os.path.exists(directory):
        os.makedirs(directory, exist_ok=True)
    with open(path, "w", encoding="utf-8") as handler:
        json.dump(state, handler, ensure_ascii=False, indent=2)


def normalize_value(value: Any) -> Any:
    if isinstance(value, (datetime, date)):
        return value.isoformat()
    if isinstance(value, Decimal):
        return float(value)
    if isinstance(value, bytes):
        return value.decode("utf-8", errors="ignore")
    return value


def pick_timestamp(row: Dict[str, Any], preferred: List[str]) -> Any:
    for key in preferred:
        if key in row:
            return row[key]
    for fallback in ("create_time", "created_at", "log_time", "timestamp", "event_time"):
        if fallback in row:
            return row[fallback]
    return None


def fetch_rows(connection: pymysql.connections.Connection, table: Dict[str, Any], last_id: int, limit: int) -> List[Dict[str, Any]]:
    column = table.get("increment_column", "id")
    sql = f"SELECT * FROM `{table['name']}` WHERE `{column}` > %s ORDER BY `{column}` ASC LIMIT %s"
    with connection.cursor(pymysql.cursors.DictCursor) as cursor:
        cursor.execute(sql, (last_id, limit))
        rows = cursor.fetchall()
    return rows


def build_payload(source_host: str, table_alias: str, rows: List[Dict[str, Any]], table_cfg: Dict[str, Any]) -> Tuple[List[Dict[str, Any]], int]:
    increment_column = table_cfg.get("increment_column", "id")
    timestamp_fields = table_cfg.get("timestamp_fields", [])
    payload_rows: List[Dict[str, Any]] = []
    max_id = 0

    for row in rows:
        remote_id = int(normalize_value(row[increment_column]))
        max_id = max(max_id, remote_id)

        data = {}
        for key, value in row.items():
            if key == "id":
                continue
            data[key] = normalize_value(value)

        created_at = pick_timestamp(data, timestamp_fields)
        payload_rows.append(
            {
                "remote_id": remote_id,
                "created_at": normalize_value(created_at) if created_at else None,
                "data": data,
            }
        )

    payload = {
        "source_host": source_host,
        "table": table_alias,
        "rows": payload_rows,
    }

    return payload, max_id


def push_payload(central_url: str, token: str, payload: Dict[str, Any], timeout: int) -> None:
    headers = {"Content-Type": "application/json", "X-SYNC-TOKEN": token}
    resp = requests.post(central_url, headers=headers, json=payload, timeout=timeout)
    if resp.status_code >= 300:
        raise RuntimeError(f"central responded {resp.status_code}: {resp.text}")


def run_sync_once(config_data: Dict[str, Any]) -> bool:
    state_file = config_data.get("state_file", "./sync_state.json")
    state = load_state(state_file)
    mysql_cfg = config_data["mysql"]

    changed = False
    connection = pymysql.connect(
        host=mysql_cfg.get("host", "127.0.0.1"),
        port=mysql_cfg.get("port", 3306),
        user=mysql_cfg.get("user", "root"),
        password=mysql_cfg.get("password", ""),
        database=mysql_cfg.get("database", "andorralee"),
        cursorclass=pymysql.cursors.DictCursor,
        charset="utf8mb4",
    )

    try:
        for table_cfg in config_data.get("tables", []):
            table_name = table_cfg["name"]
            table_alias = table_cfg.get("remote_alias", table_name)
            last_id = int(state.get(table_name, 0))
            rows = fetch_rows(connection, table_cfg, last_id, config_data.get("batch_size", 500))
            if not rows:
                LOGGER.debug("table %s no new rows after id %s", table_name, last_id)
                continue

            payload, max_id = build_payload(config_data["source_host"], table_alias, rows, table_cfg)
            push_payload(config_data["central_url"], config_data["sync_token"], payload, config_data.get("timeout", 10))
            state[table_name] = max_id
            changed = True
            LOGGER.info("pushed %s rows from %s (max id %s)", len(rows), table_name, max_id)
    finally:
        connection.close()

    if changed:
        save_state(state_file, state)
    return changed


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Incremental sync client")
    parser.add_argument("--config", default="tools/remote_sync_config.json", help="Path to config JSON")
    parser.add_argument("--interval", type=int, default=60, help="Loop interval seconds when --loop set")
    parser.add_argument("--loop", action="store_true", help="Run continuously")
    parser.add_argument("--log-level", default="INFO", help="Logging level")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    logging.basicConfig(level=getattr(logging, args.log_level.upper(), logging.INFO), format="%(asctime)s [%(levelname)s] %(message)s")

    if not os.path.exists(args.config):
        LOGGER.error("config file %s not found", args.config)
        sys.exit(1)

    config_data = load_config(args.config)

    while True:
        try:
            run_sync_once(config_data)
        except Exception as exc:  # pylint: disable=broad-except
            LOGGER.error("sync failed: %s", exc)
            if not args.loop:
                raise
        if not args.loop:
            break
        time.sleep(max(5, args.interval))

if __name__ == "__main__":
    main()
