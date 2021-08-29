package config

import "testing"

func TestLoad(t *testing.T) {
	err := Load()
	if err != nil {
		t.Error(err)
	}
}
