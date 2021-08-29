package config

import (
	"encoding/json"
	"os"
)

var C Config

type (
	Config struct {
		Addr
		Listen                 string
		HTTPListen             string
		ExecuteBeforeUseNewSQL string
	}

	Addr struct {
		Old    []string
		New    []string
		Client []string
	}
)

func Load() error {
	return LoadPath("./config.json")
}

func LoadPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(f)

	err = dec.Decode(&C)
	if err != nil {
		return err
	}

	return nil
}
