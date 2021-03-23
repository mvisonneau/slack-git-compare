package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Format represents the format of the config file
type Format uint8

const (
	// FormatJSON represents a Config written in json format
	FormatJSON Format = iota

	// FormatYAML represents a Config written in yaml format
	FormatYAML
)

// ParseFile reads the content of a file and attempt to unmarshal it
// into a Config
func ParseFile(filename string) (c Config, err error) {
	var t Format
	var fileBytes []byte

	// Figure out what type of config file we provided
	t, err = GetTypeFromFileExtension(filename)
	if err != nil {
		return
	}

	// Read the content of the config file
	fileBytes, err = ioutil.ReadFile(filepath.Clean(filename))
	if err != nil {
		return
	}

	// Parse the content and return Config
	return Parse(t, fileBytes)
}

// Parse unmarshal provided bytes with given ConfigType into a Config object
func Parse(f Format, bytes []byte) (Config, error) {
	cfg := NewConfig()
	var err error

	switch f {
	case FormatJSON:
		err = json.Unmarshal(bytes, &cfg)
	case FormatYAML:
		err = yaml.Unmarshal(bytes, &cfg)
	default:
		err = fmt.Errorf("unsupported config type '%+v'", f)
	}
	return cfg, err
}

// GetTypeFromFileExtension returns the ConfigType based upon the extension of
// the file
func GetTypeFromFileExtension(filename string) (f Format, err error) {
	ext := filepath.Ext(filename)
	switch ext {
	case ".json":
		f = FormatJSON
	case ".yml", ".yaml":
		f = FormatYAML
	default:
		err = fmt.Errorf("unsupported config type '%s', expected .dhall, .json or .y(a)ml", ext)
	}
	return
}
