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
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFromFile(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test-config.yaml")

	configContent := `accountNumber: A-12345678
meterPointID: "1000000000000"
apiKey: sk_live_test_key
outputFile: test_output.json`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	config := &Config{}
	err = loadConfigFromFile(configFile, config)
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}

	// Verify config values
	if config.AccountNumber != "A-12345678" {
		t.Errorf("Expected account number 'A-12345678', got '%s'", config.AccountNumber)
	}
	if config.MeterPointID != "1000000000000" {
		t.Errorf("Expected meter point ID '1000000000000', got '%s'", config.MeterPointID)
	}
	if config.APIKey != "sk_live_test_key" {
		t.Errorf("Expected API key 'sk_live_test_key', got '%s'", config.APIKey)
	}
	if config.OutputFile != "test_output.json" {
		t.Errorf("Expected output file 'test_output.json', got '%s'", config.OutputFile)
	}
}

func TestLoadConfigFromFile_InvalidFile(t *testing.T) {
	config := &Config{}
	err := loadConfigFromFile("/nonexistent/file.yaml", config)
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestLoadConfigFromFile_InvalidYAML(t *testing.T) {
	// Create a temporary file with invalid YAML
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "invalid-config.yaml")

	invalidContent := `invalid: yaml: content:
	- with
	- bad: formatting
	[and brackets`

	err := os.WriteFile(configFile, []byte(invalidContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	config := &Config{}
	err = loadConfigFromFile(configFile, config)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_EnvironmentVariables(t *testing.T) {
	// Set up environment variables
	originalAccount := os.Getenv("ACCOUNT_NUMBER")
	originalMeter := os.Getenv("METER_POINT_ID")
	originalAPIKey := os.Getenv("OCTOPUS_API_KEY")

	defer func() {
		os.Setenv("ACCOUNT_NUMBER", originalAccount)
		os.Setenv("METER_POINT_ID", originalMeter)
		os.Setenv("OCTOPUS_API_KEY", originalAPIKey)
	}()

	// Clear any existing values
	os.Unsetenv("ACCOUNT_NUMBER")
	os.Unsetenv("METER_POINT_ID")
	os.Unsetenv("OCTOPUS_API_KEY")

	// Test with environment variables
	os.Setenv("ACCOUNT_NUMBER", "A-99999999")
	os.Setenv("METER_POINT_ID", "2000000000000")
	os.Setenv("OCTOPUS_API_KEY", "sk_live_env_key")

	// Reset flag values to empty
	originalAccountFlag := *accountNumber
	originalMeterFlag := *meterPointID
	originalAPIKeyFlag := *apiKey
	defer func() {
		*accountNumber = originalAccountFlag
		*meterPointID = originalMeterFlag
		*apiKey = originalAPIKeyFlag
	}()

	*accountNumber = ""
	*meterPointID = ""
	*apiKey = ""

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.AccountNumber != "A-99999999" {
		t.Errorf("Expected account number from env 'A-99999999', got '%s'", config.AccountNumber)
	}
	if config.MeterPointID != "2000000000000" {
		t.Errorf("Expected meter point ID from env '2000000000000', got '%s'", config.MeterPointID)
	}
	if config.APIKey != "sk_live_env_key" {
		t.Errorf("Expected API key from env 'sk_live_env_key', got '%s'", config.APIKey)
	}
}

func TestLoadConfig_CommandLineFlags(t *testing.T) {
	// Reset flag values
	originalAccountFlag := *accountNumber
	originalMeterFlag := *meterPointID
	originalAPIKeyFlag := *apiKey
	originalOutputFlag := *outputFile

	defer func() {
		*accountNumber = originalAccountFlag
		*meterPointID = originalMeterFlag
		*apiKey = originalAPIKeyFlag
		*outputFile = originalOutputFlag
	}()

	// Set flag values
	*accountNumber = "A-88888888"
	*meterPointID = "3000000000000"
	*apiKey = "sk_live_flag_key"
	*outputFile = "flag_output.json"

	// Clear environment variables to ensure flags take precedence
	originalAccount := os.Getenv("ACCOUNT_NUMBER")
	originalMeter := os.Getenv("METER_POINT_ID")
	originalAPIKey := os.Getenv("OCTOPUS_API_KEY")

	os.Unsetenv("ACCOUNT_NUMBER")
	os.Unsetenv("METER_POINT_ID")
	os.Unsetenv("OCTOPUS_API_KEY")

	defer func() {
		os.Setenv("ACCOUNT_NUMBER", originalAccount)
		os.Setenv("METER_POINT_ID", originalMeter)
		os.Setenv("OCTOPUS_API_KEY", originalAPIKey)
	}()

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.AccountNumber != "A-88888888" {
		t.Errorf("Expected account number from flag 'A-88888888', got '%s'", config.AccountNumber)
	}
	if config.MeterPointID != "3000000000000" {
		t.Errorf("Expected meter point ID from flag '3000000000000', got '%s'", config.MeterPointID)
	}
	if config.APIKey != "sk_live_flag_key" {
		t.Errorf("Expected API key from flag 'sk_live_flag_key', got '%s'", config.APIKey)
	}
	if config.OutputFile != "flag_output.json" {
		t.Errorf("Expected output file from flag 'flag_output.json', got '%s'", config.OutputFile)
	}
}

func TestLoadConfig_MissingAPIKey(t *testing.T) {
	// Clear all configuration sources
	originalAccountFlag := *accountNumber
	originalMeterFlag := *meterPointID
	originalAPIKeyFlag := *apiKey

	*accountNumber = ""
	*meterPointID = ""
	*apiKey = ""

	originalAccount := os.Getenv("ACCOUNT_NUMBER")
	originalMeter := os.Getenv("METER_POINT_ID")
	originalAPIKey := os.Getenv("OCTOPUS_API_KEY")

	os.Unsetenv("ACCOUNT_NUMBER")
	os.Unsetenv("METER_POINT_ID")
	os.Unsetenv("OCTOPUS_API_KEY")

	defer func() {
		*accountNumber = originalAccountFlag
		*meterPointID = originalMeterFlag
		*apiKey = originalAPIKeyFlag
		os.Setenv("ACCOUNT_NUMBER", originalAccount)
		os.Setenv("METER_POINT_ID", originalMeter)
		os.Setenv("OCTOPUS_API_KEY", originalAPIKey)
	}()

	_, err := loadConfig()
	if err == nil {
		t.Error("Expected error for missing API key, got nil")
	}

	expectedError := "API key is required (use -key flag, config file, or OCTOPUS_API_KEY env var)"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestLoadConfig_MissingAccountNumber(t *testing.T) {
	// Set API key but clear account number
	originalAPIKeyFlag := *apiKey
	originalAccountFlag := *accountNumber
	originalMeterFlag := *meterPointID

	*apiKey = "sk_live_test_key"
	*accountNumber = ""
	*meterPointID = ""

	originalAPIKey := os.Getenv("OCTOPUS_API_KEY")
	originalAccount := os.Getenv("ACCOUNT_NUMBER")
	originalMeter := os.Getenv("METER_POINT_ID")

	os.Setenv("OCTOPUS_API_KEY", "sk_live_test_key")
	os.Unsetenv("ACCOUNT_NUMBER")
	os.Unsetenv("METER_POINT_ID")

	defer func() {
		*apiKey = originalAPIKeyFlag
		*accountNumber = originalAccountFlag
		*meterPointID = originalMeterFlag
		os.Setenv("OCTOPUS_API_KEY", originalAPIKey)
		os.Setenv("ACCOUNT_NUMBER", originalAccount)
		os.Setenv("METER_POINT_ID", originalMeter)
	}()

	_, err := loadConfig()
	if err == nil {
		t.Error("Expected error for missing account number, got nil")
	}

	expectedError := "account number is required (use -account flag, config file, or ACCOUNT_NUMBER env var)"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestLoadConfig_MissingMeterPointID(t *testing.T) {
	// Set API key and account number but clear meter point ID
	originalAPIKeyFlag := *apiKey
	originalAccountFlag := *accountNumber
	originalMeterFlag := *meterPointID

	*apiKey = "sk_live_test_key"
	*accountNumber = "A-12345678"
	*meterPointID = ""

	originalAPIKey := os.Getenv("OCTOPUS_API_KEY")
	originalAccount := os.Getenv("ACCOUNT_NUMBER")
	originalMeter := os.Getenv("METER_POINT_ID")

	os.Setenv("OCTOPUS_API_KEY", "sk_live_test_key")
	os.Setenv("ACCOUNT_NUMBER", "A-12345678")
	os.Unsetenv("METER_POINT_ID")

	defer func() {
		*apiKey = originalAPIKeyFlag
		*accountNumber = originalAccountFlag
		*meterPointID = originalMeterFlag
		os.Setenv("OCTOPUS_API_KEY", originalAPIKey)
		os.Setenv("ACCOUNT_NUMBER", originalAccount)
		os.Setenv("METER_POINT_ID", originalMeter)
	}()

	_, err := loadConfig()
	if err == nil {
		t.Error("Expected error for missing meter point ID, got nil")
	}

	expectedError := "meter point ID is required (use -meter flag, config file, or METER_POINT_ID env var)"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}
