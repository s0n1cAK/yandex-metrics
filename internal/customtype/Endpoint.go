package customtype

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type Endpoint string

func formatEndpoint(value string) (Endpoint, error) {
	if !strings.Contains(value, "://") {
		value = "http://" + value
	}

	u, err := url.Parse(value)
	if err != nil || u.Host == "" {
		return "", errors.New("invalid endpoint format, must be scheme://host:port")
	}

	hostPort := strings.Split(u.Host, ":")
	if len(hostPort) != 2 {
		return "", errors.New("endpoint must include port")
	}

	port, err := strconv.Atoi(hostPort[1])
	if err != nil || port < 1 || port > 65535 {
		return "", errors.New("invalid port number")
	}

	return Endpoint(value), err
}

func (e *Endpoint) String() string {
	return string(*e)
}

func (e *Endpoint) Type() string {
	return "endpoint"
}

func (e *Endpoint) Set(value string) error {
	gValue, err := formatEndpoint(string(value[:]))
	if err != nil {
		return err
	}
	*e = gValue
	return nil
}

func (e *Endpoint) UnmarshalText(text []byte) error {
	gValue, err := formatEndpoint(string(text[:]))
	if err != nil {
		return err
	}
	*e = gValue
	return nil
}

func (e *Endpoint) HostPort() string {
	u, err := url.Parse(string(*e))
	if err != nil {
		return string(*e)
	}
	return u.Host
}
