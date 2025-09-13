package utils

import (
	"fmt"
	"strings"
)

// ParseMySQLDSN simplified parser (use full lib in prod).
func ParseMySQLDSN(dsn string) (map[string]string, error) {
	parts := strings.Split(dsn, "://")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid DSN")
	}
	userPass := strings.Split(parts[1], "@")
	if len(userPass) < 2 {
		return nil, fmt.Errorf("invalid user@pass")
	}
	hostPortDB := strings.Split(userPass[1], "/")
	hostPort := strings.Split(hostPortDB[0], ":")
	return map[string]string{
		"user": strings.Split(userPass[0], ":")[0],
		"pass": strings.Split(userPass[0], ":")[1],
		"host": hostPort[0],
		"port": hostPort[1],
		"db":   hostPortDB[1],
	}, nil
}
