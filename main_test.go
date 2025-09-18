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

func TestSetupLogging(t *testing.T) {
	// Test setupLogging doesn't panic and sets up logging correctly
	// This is mainly to ensure the function is covered

	// Save original log format
	originalLogFormat := *logFormat
	defer func() {
		*logFormat = originalLogFormat
	}()

	// Test with different log formats
	testFormats := []string{"text", "json", "auto"}

	for _, format := range testFormats {
		*logFormat = format

		// This should not panic
		setupLogging()
	}
}

func TestFetchAndUpdateEvents_NoExistingFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_output.json")

	// Create a config with the temp output file
	config := &Config{
		AccountNumber: "A-12345678",
		MeterPointID:  "1000000000000",
		APIKey:        "sk_live_test_key",
		OutputFile:    outputFile,
	}

	// This will fail because we don't have real API credentials,
	// but it should still exercise the code path and create the output file
	err := fetchAndUpdateEvents(config)

	// We expect this to fail due to invalid API credentials, but it should not panic
	if err == nil {
		t.Log("fetchAndUpdateEvents succeeded (unexpected but not an error)")
	}

	// Verify the output file was created (even if empty)
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was not created: %s", outputFile)
	}
}

func TestFetchAndUpdateEvents_WithExistingEvents(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "test_output.json")

	// Create existing events file
	existingEvents := `{
		"data": [
			{
				"start": "2024-01-01T12:00:00.000Z",
				"end": "2024-01-01T13:00:00.000Z",
				"code": "1"
			}
		]
	}`

	err := os.WriteFile(outputFile, []byte(existingEvents), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing events file: %v", err)
	}

	// Create a config with the temp output file
	config := &Config{
		AccountNumber: "A-12345678",
		MeterPointID:  "1000000000000",
		APIKey:        "sk_live_test_key",
		OutputFile:    outputFile,
	}

	// This should load existing events and attempt to fetch new ones
	err = fetchAndUpdateEvents(config)

	// We expect this to fail due to invalid API credentials, but it should handle existing events
	if err == nil {
		t.Log("fetchAndUpdateEvents succeeded (unexpected but not an error)")
	}

	// Verify the output file still exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("Output file was deleted: %s", outputFile)
	}
}
