package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ConfigType uint8

const (
	ConfigTypeJson ConfigType = iota
	ConfigTypeYaml
)

func ParseFile(filename string) (c Config, err error) {
	var ct ConfigType
	var fileBytes []byte

	// Figure out what type of config file we provided
	ct, err = GetConfigTypeFromFileExtension(filename)
	if err != nil {
		return
	}

	// Read the content of the config file
	fileBytes, err = ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return
	}

	// Parse the content and return Config
	return Parse(ct, fileBytes)
}

func Parse(ct ConfigType, bytes []byte) (Config, error) {
	cfg := NewConfig()
	var err error

	switch ct {
	case ConfigTypeJson:
		err = json.Unmarshal(bytes, &cfg)
	case ConfigTypeYaml:
		err = yaml.Unmarshal(bytes, &cfg)
	default:
		err = fmt.Errorf("unsupported config type '%+v'", ct)
	}
	return cfg, err
}

func GetConfigTypeFromFileExtension(filename string) (c ConfigType, err error) {
	ext := filepath.Ext(filename)
	switch ext {
	case ".yml", ".yaml":
		c = ConfigTypeYaml
	case ".json":
		c = ConfigTypeJson
	default:
		err = fmt.Errorf("unsupported config type '%s', expected .dhall, .json or .y(a)ml", ext)
	}
	return
}
