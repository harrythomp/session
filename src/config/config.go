package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"harry/session/src/session"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Location     string
	SearchPaths  []string          `json:"searchPaths"`
	IncludePaths []string          `json:"includePaths"`
	Favourites   map[string]string `json:"favourites"`
}

const defaultConfigContent = `{
	"searchPaths": [
	],
	"includePaths": [
	],
	"favourites": {
	}
}`

func ParseFromConfigDir(location string) (Config, error) {
	file, err := os.Open(filepath.Join(location, "config.json"))
	if errors.Is(err, os.ErrNotExist) {
		return makeDefaultConfig(location)
	}
	if err != nil {
		return Config{}, fmt.Errorf("error when opening config file: %w", err)
	}
	defer file.Close()
	conf, err := parseConfig(file)
	conf.Location = location
	for key, favourite := range conf.Favourites {
		expanded, err := session.ExpandPathHomeDir(favourite)
		if err != nil {
			return Config{}, err
		}
		conf.Favourites[key] = expanded
	}
	return conf, err
}

func makeDefaultConfig(location string) (Config, error) {
	err := os.MkdirAll(location, 0755) // u: rwx, g: r-x, o: r-x
	if err != nil {
		return Config{}, fmt.Errorf("error when creating config dir: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(location, "config.json"), os.O_CREATE|os.O_WRONLY, 0644) // u: rw-, g: r--, o: r--
	if err != nil {
		return Config{}, fmt.Errorf("error when creating config file: %w", err)
	}
	defer file.Close()
	_, err = file.WriteString(defaultConfigContent)
	if err != nil {
		return Config{}, fmt.Errorf("error when writing default config: %w", err)
	}
	return parseConfig(strings.NewReader(defaultConfigContent))
}

func parseConfig(r io.Reader) (Config, error) {
	var conf Config
	decoder := json.NewDecoder(r)
	err := decoder.Decode(&conf)
	if err != nil {
		return Config{}, fmt.Errorf("error when decoding config: %w", err)
	}
	return conf, nil
}
