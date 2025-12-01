package handlers

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	portMappingFormatHostFirst      = "host-first"
	portMappingFormatContainerFirst = "container-first"
)

// normalizeRequestPortMappings inspects request metadata and normalizes port mappings into
// container-port-keyed format when the client submits host-port-keyed payloads.
func normalizeRequestPortMappings(c *gin.Context, description string, mappings map[string]string) (map[string]string, bool) {
	formatHint := detectPortMappingFormat(description, c.GetHeader("X-Port-Mapping-Format"))
	return normalizePortMappingsByFormat(mappings, formatHint)
}

func detectPortMappingFormat(description, headerValue string) string {
	if format := parsePortMappingFormat(headerValue); format != "" {
		return format
	}
	if format := parsePortMappingFormat(os.Getenv("PORT_MAPPING_INPUT_FORMAT")); format != "" {
		return format
	}

	desc := strings.ToLower(description)
	if desc != "" && strings.Contains(desc, "web ui") {
		return portMappingFormatHostFirst
	}

	return portMappingFormatContainerFirst
}

func parsePortMappingFormat(value string) string {
	if value == "" {
		return ""
	}

	switch strings.ToLower(strings.TrimSpace(value)) {
	case "host", "host-first", "host_key", "host-key", "host-keyed", "hostport", "host_port_first", "host-first-format":
		return portMappingFormatHostFirst
	case "container", "container-first", "container-keyed", "container_port_first", "container-first-format":
		return portMappingFormatContainerFirst
	default:
		return ""
	}
}

func normalizePortMappingsByFormat(mappings map[string]string, format string) (map[string]string, bool) {
	if len(mappings) == 0 {
		return mappings, false
	}

	if strings.EqualFold(format, portMappingFormatHostFirst) {
		normalized := make(map[string]string, len(mappings))
		for hostPort, containerPort := range mappings {
			normalized[strings.TrimSpace(containerPort)] = strings.TrimSpace(hostPort)
		}

		cleaned := make(map[string]string, len(normalized))
		for containerPort, hostPort := range normalized {
			if containerPort == "" {
				continue
			}
			cleaned[containerPort] = hostPort
		}

		if len(cleaned) == 0 {
			return mappings, false
		}
		return cleaned, true
	}

	return mappings, false
}
