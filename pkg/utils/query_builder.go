package utils

import (
	"fmt"
	"strings"
)

// PageQuery 保存分页与排序参数
type PageQuery struct {
	Page    int
	Size    int
	OrderBy string // e.g. "created_at desc"
}

// BuildLimitOffset 生成跨 MySQL / 达梦通用分页片段 (返回 SQL 末尾附加部分)
// dm (达梦) 支持 ANSI OFFSET/FETCH，自 v8 起可用
func BuildLimitOffset(dbMode string, pq PageQuery) string {
	if pq.Size <= 0 { return "" }
	if pq.Page <= 0 { pq.Page = 1 }
	offset := (pq.Page - 1) * pq.Size
	suffix := ""
	if pq.OrderBy != "" {
		suffix += " ORDER BY " + sanitizeOrderBy(pq.OrderBy)
	}
	if strings.ToLower(dbMode) == "dameng" {
		// 达梦: OFFSET <n> ROWS FETCH NEXT <m> ROWS ONLY
		suffix += fmt.Sprintf(" OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, pq.Size)
	} else {
		// MySQL
		suffix += fmt.Sprintf(" LIMIT %d OFFSET %d", pq.Size, offset)
	}
	return suffix
}

// BuildDateEqualSameDay 生成同日匹配表达式 (列名必须是时间/日期型)
// 返回 WHERE 条件片段与参数 (MySQL 使用 DATE(col)=?; 达梦使用 TO_CHAR(col,'YYYY-MM-DD')=?)
func BuildDateEqualSameDay(dbMode, column string) (string, string) {
	if strings.ToLower(dbMode) == "dameng" {
		return fmt.Sprintf("TO_CHAR(%s,'YYYY-MM-DD') = ?", column), "" // 占位，调用方填日期字符串
	}
	return fmt.Sprintf("DATE(%s) = ?", column), ""
}

// sanitizeOrderBy 简单清理 OrderBy，防注入（允许字母数字下划线逗号空格及 ASC/DESC）
func sanitizeOrderBy(ob string) string {
	allowed := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_ ,.`")
	var b strings.Builder
	for _, r := range ob {
		keep := false
		for _, a := range allowed { if r == a { keep = true; break } }
		if keep { b.WriteRune(r) }
	}
	return b.String()
}
