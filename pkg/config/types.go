package config

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Gituser143/stunning-octo-enigma/pkg/trigger"
)

// Host holds details for an endpoint.
type Host struct {
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Scheme string `json:"scheme"`
}

// LoadParameters hold  values to be used when load testing.
type LoadParameters struct {
	DistributionType string `json:"distributionType"`
	Steps            int    `json:"steps"`
	Duration         int    `json:"duration"`
	Workers          int    `json:"workers"`
	MinRate          int    `json:"minRate"`
	MaxRate          int    `json:"maxRate"`
}

// Config holds configuration details of Kiali, Application endpoints along
// with relevant load parameters, namespaces to use and per deployment
// resource thresholds.
type Config struct {
	KialiHost  Host               `json:"kialiHost"`
	AppHost    Host               `json:"appHost"`
	Thresholds trigger.Thresholds `json:"thresholds"`
	LoadConfig LoadParameters     `json:"loadParameters"`
	Namespaces []string           `json:"namespaces"`
}

// ErrInavlidConfigPath signifies the error when a path to a config file is not
// specified correctly
var ErrInavlidConfigPath = fmt.Errorf("invalid path specified for config file")

// GetConfig return a config variable and a nil error on succesfull parsing of
// a config file given by "path". On error, it returns an empty config and
// the corresponding error
func GetConfig(path string) (Config, error) {
	// Init Config
	c := Config{}

	// Ensure valid file
	fileInfo, err := os.Stat(path)
	if err != nil {
		return c, ErrInavlidConfigPath
	}

	// TODO: Handle multiple file configs cleanly
	if fileInfo.IsDir() {
		// Walk through files in directory
		err = filepath.Walk(path,
			func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Filter on JSON files
				if !info.IsDir() && strings.HasSuffix(path, ".json") {
					bs, err := ioutil.ReadFile(path)
					if err != nil {
						return err
					}
					err = json.Unmarshal(bs, &c)
					if err != nil {
						return err
					}
				}

				return nil
			})

		if err != nil {
			return c, err
		}
	} else {
		// Read file as byte slice
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return c, err
		}

		// Unmarshal file into config
		err = json.Unmarshal(bs, &c)
		if err != nil {
			return c, err
		}
	}

	return c, nil
}
