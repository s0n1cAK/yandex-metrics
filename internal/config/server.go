package config

import (
	"errors"
	"flag"
	"strconv"
	"strings"
)

type ServerConfig struct {
	Address string
	Port    int
}

func (o ServerConfig) String() string {
	return o.Address + ":" + strconv.Itoa(o.Port)
}

func (o *ServerConfig) Set(s string) error {
	gAddress := strings.Split(s, ":")
	if len(gAddress) < 2 {
		return errors.New("Need address in a form host:port")
	}

	port, err := strconv.Atoi(gAddress[1])
	if err != nil {
		return err
	}

	o.Address = gAddress[0]
	o.Port = port
	return nil
}

func NewServerConfig() *ServerConfig {
	addr := &ServerConfig{
		Address: "localhost",
		Port:    8080,
	}

	flag.Var(addr, "a", "Server of address in format host:port")
	flag.Parse()

	return addr
}
