/*
 * Copyright 2025 Matthew Gall <me@matthewgall.dev>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	AccountNumber string `yaml:"accountNumber"`
	MeterPointID  string `yaml:"meterPointID"`
	APIKey        string `yaml:"apiKey"`
	OutputFile    string `yaml:"outputFile"`
}

var (
	configFile    = flag.String("config", "", "Path to configuration file")
	accountNumber = flag.String("account", "", "Octopus Energy Account Number")
	meterPointID  = flag.String("meter", "", "Meter Point ID (MPAN)")
	apiKey        = flag.String("key", "", "Octopus Energy API Key")
	outputFile    = flag.String("output", "free_electricity.json", "Output file path")
	version       = flag.Bool("version", false, "Show version information")
)

func loadConfig() (*Config, error) {
	flag.Parse()

	if *version {
		fmt.Println("octoevents v1.0.0")
		os.Exit(0)
	}

	config := &Config{
		OutputFile: *outputFile,
	}

	if *configFile != "" {
		if err := loadConfigFromFile(*configFile, config); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	if *accountNumber != "" {
		config.AccountNumber = *accountNumber
	} else if config.AccountNumber == "" {
		config.AccountNumber = getEnvOrDefault("ACCOUNT_NUMBER", "")
	}

	if *meterPointID != "" {
		config.MeterPointID = *meterPointID
	} else if config.MeterPointID == "" {
		config.MeterPointID = getEnvOrDefault("METER_POINT_ID", "")
	}

	if *apiKey != "" {
		config.APIKey = *apiKey
	} else if config.APIKey == "" {
		config.APIKey = os.Getenv("OCTOPUS_API_KEY")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required (use -key flag, config file, or OCTOPUS_API_KEY env var)")
	}

	if config.AccountNumber == "" {
		return nil, fmt.Errorf("account number is required (use -account flag, config file, or ACCOUNT_NUMBER env var)")
	}

	if config.MeterPointID == "" {
		return nil, fmt.Errorf("meter point ID is required (use -meter flag, config file, or METER_POINT_ID env var)")
	}

	return config, nil
}

func loadConfigFromFile(filename string, config *Config) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}

	slog.Info("Loaded configuration from file", "filename", filename)
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}