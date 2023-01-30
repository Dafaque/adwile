package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DbEnableStdout       bool        `json:"db_enable_stdout"`
	DbOpTimeoutSec       int         `json:"db_op_timeout_sec"`
	ConnStr              string      `json:"conn_str"`
	CheckTimeoutSec      int         `json:"check_timeout_sec"`
	HttpClientTimeoutSec int         `json:"http_client_timeout_sec"`
	Urls                 []ConfigUrl `json:"urls"`
}

type ConfigUrl struct {
	Url            string   `json:"url"`
	Checks         []string `json:"checks"`
	MinChecksCount int      `json:"min_checks_cnt"`
}

func NewConfig(path string) (*Config, error) {
	data, errReadFile := os.ReadFile(path)
	if errReadFile != nil {
		return nil, errReadFile
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
