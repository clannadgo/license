package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	PrivateKeyPath   string `json:"privateKeyPath"`
	PublicKeyPath    string `json:"publicKeyPath"`
	LicenseStorePath string `json:"licenseStorePath"`
}

var Conf *Config

// 加载配置
func LoadConfig() error {
	if Conf != nil {
		return nil
	}
	var cfg *Config
	f, err := os.ReadFile("./config.json")
	if err != nil {
		return err
	}
	if err := json.Unmarshal(f, &cfg); err != nil {

		return err
	}
	Conf = cfg
	return nil
}

func init() {
	if err := LoadConfig(); err != nil {
		panic(err)
	}
}
