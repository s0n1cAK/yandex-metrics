package config

import (
	"errors"
	"flag"
	"os"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

var (
	ErrInvalidAddressFormat = errors.New("need address in a form host:port")
)

type ServerConfig struct {
	Address string
	Port    int
	Logger  *zap.Logger
}

func (o ServerConfig) String() string {
	return o.Address + ":" + strconv.Itoa(o.Port)
}

func (o *ServerConfig) Set(s string) error {
	gAddress := strings.Split(s, ":")
	if len(gAddress) < 2 {
		return ErrInvalidAddressFormat
	}

	port, err := strconv.Atoi(gAddress[1])
	if err != nil {
		return err
	}

	o.Address = gAddress[0]
	o.Port = port
	return nil
}

func (o *ServerConfig) ParseENV() error {
	if env, exist := os.LookupEnv("ADDRESS"); exist {
		gAddress := strings.Split(env, ":")
		if len(gAddress) < 2 {
			return ErrInvalidAddressFormat
		}

		port, err := strconv.Atoi(gAddress[1])
		if err != nil {
			return err
		}

		o.Address = gAddress[0]
		o.Port = port
		return nil
	}
	return nil
}

func NewServerConfig(log *zap.Logger) (*ServerConfig, error) {
	addr := &ServerConfig{
		Address: "localhost",
		Port:    8080,
		Logger:  log,
	}

	err := addr.ParseENV()
	if err != nil {
		return &ServerConfig{}, err
	}

	flag.Var(addr, "a", "Server of address in format host:port")
	flag.Parse()

	return addr, nil
}
