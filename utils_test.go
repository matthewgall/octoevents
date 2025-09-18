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
	"runtime/debug"
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	// Test with default values (should return "dev")
	originalBuildVersion := buildVersion
	originalBuildCommit := buildCommit
	defer func() {
		buildVersion = originalBuildVersion
		buildCommit = originalBuildCommit
	}()

	// Test default case
	buildVersion = "dev"
	buildCommit = "unknown"
	version := GetVersion()
	if version != "dev" {
		t.Errorf("Expected 'dev', got '%s'", version)
	}

	// Test with explicit version
	buildVersion = "1.0.0"
	version = GetVersion()
	if version != "1.0.0" {
		t.Errorf("Expected '1.0.0', got '%s'", version)
	}

	// Test with build info fallback
	buildVersion = "dev"
	buildCommit = "unknown"

	// Mock build info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		// Test VCS revision fallback
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" {
				version := GetVersion()
				if version == "dev" {
					// If no VCS info, should still return dev
					break
				}
				// If VCS info exists, should use it
				if len(version) < 7 {
					t.Errorf("Expected at least 7 characters for commit hash, got '%s'", version)
				}
				break
			}
		}
	}
}

func TestGetUserAgent(t *testing.T) {
	userAgent := GetUserAgent()
	expectedPrefix := "matthewgall/octoevents/"

	if !strings.HasPrefix(userAgent, expectedPrefix) {
		t.Errorf("Expected user agent to start with '%s', got '%s'", expectedPrefix, userAgent)
	}

	// Should contain version information
	if !strings.Contains(userAgent, "/") {
		t.Errorf("Expected user agent to contain version separator '/', got '%s'", userAgent)
	}
}

func TestDetectLogFormat(t *testing.T) {
	// Test GitHub Actions environment
	originalCI := os.Getenv("CI")
	originalGHActions := os.Getenv("GITHUB_ACTIONS")
	originalK8s := os.Getenv("KUBERNETES_SERVICE_HOST")
	defer func() {
		os.Setenv("CI", originalCI)
		os.Setenv("GITHUB_ACTIONS", originalGHActions)
		os.Setenv("KUBERNETES_SERVICE_HOST", originalK8s)
	}()

	// Test GitHub Actions
	os.Setenv("GITHUB_ACTIONS", "true")
	format := detectLogFormat()
	if format != "json" {
		t.Errorf("Expected 'json' for GitHub Actions, got '%s'", format)
	}

	// Test CI environment
	os.Setenv("GITHUB_ACTIONS", "")
	os.Setenv("CI", "true")
	format = detectLogFormat()
	if format != "json" {
		t.Errorf("Expected 'json' for CI, got '%s'", format)
	}

	// Test Kubernetes environment
	os.Setenv("CI", "")
	os.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
	format = detectLogFormat()
	if format != "json" {
		t.Errorf("Expected 'json' for Kubernetes, got '%s'", format)
	}

	// Test local development (default)
	os.Setenv("KUBERNETES_SERVICE_HOST", "")
	format = detectLogFormat()
	if format != "text" {
		t.Errorf("Expected 'text' for local development, got '%s'", format)
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	key := "TEST_ENV_VAR"
	defaultValue := "default_value"

	// Test with environment variable set
	os.Setenv(key, "test_value")
	defer os.Unsetenv(key)

	result := getEnvOrDefault(key, defaultValue)
	if result != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", result)
	}

	// Test with environment variable not set
	os.Unsetenv(key)
	result = getEnvOrDefault(key, defaultValue)
	if result != defaultValue {
		t.Errorf("Expected '%s', got '%s'", defaultValue, result)
	}
}
