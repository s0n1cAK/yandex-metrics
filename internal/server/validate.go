package server

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/s0n1cAK/yandex-metrics/internal/config/server"
)

func parseURL(cfg *server.Config) (string, int, error) {
	OP := "parseURl"

	parts := strings.Split(cfg.Endpoint.HostPort(), ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("%s: invalid endpoint format", OP)
	}

	domain := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, fmt.Errorf("%s: %s", OP, err)
	}

	if port <= minPort || port >= maxPort {
		return "", 0, fmt.Errorf("%s: %v is not an valid port", OP, port)
	}

	return domain, port, nil
}
