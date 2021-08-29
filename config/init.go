package config

import "log"

func init() {
	err := Load()
	if err != nil {
		log.Println(err)
	}
}
