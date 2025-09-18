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
	"time"
)

func TestCacheETag(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	etag := "test-etag-value"
	cacheETagToDir(tempDir, etag)

	// Verify ETag was cached
	cachedETag := getCachedETagFromDir(tempDir)
	if cachedETag != etag {
		t.Errorf("Expected cached ETag '%s', got '%s'", etag, cachedETag)
	}

	// Verify file exists
	etagFile := filepath.Join(tempDir, "etag")
	if _, err := os.Stat(etagFile); os.IsNotExist(err) {
		t.Error("ETag file was not created")
	}

	// Verify file contents
	data, err := os.ReadFile(etagFile)
	if err != nil {
		t.Fatalf("Failed to read ETag file: %v", err)
	}
	if string(data) != etag {
		t.Errorf("Expected file content '%s', got '%s'", etag, string(data))
	}
}

func TestGetCachedETag_NoCache(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Test with no cached ETag
	cachedETag := getCachedETagFromDir(tempDir)
	if cachedETag != "" {
		t.Errorf("Expected empty string for no cache, got '%s'", cachedETag)
	}
}

func TestCacheEvents(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	events := []Event{
		{
			Code:    "1",
			StartAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			EndAt:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			IsTest:  boolPtr(true),
		},
		{
			Code:    "2",
			StartAt: time.Date(2024, 1, 2, 12, 0, 0, 0, time.UTC),
			EndAt:   time.Date(2024, 1, 2, 13, 0, 0, 0, time.UTC),
			IsTest:  boolPtr(false),
		},
	}

	cacheEventsToDir(tempDir, events)

	// Verify events were cached
	cachedEvents, err := getCachedEventsFromDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to get cached events: %v", err)
	}

	if len(cachedEvents) != len(events) {
		t.Errorf("Expected %d cached events, got %d", len(events), len(cachedEvents))
	}

	// Verify file exists
	eventsFile := filepath.Join(tempDir, "david_events.json")
	if _, err := os.Stat(eventsFile); os.IsNotExist(err) {
		t.Error("Events cache file was not created")
	}

	// Verify file contents can be loaded
	data, err := os.ReadFile(eventsFile)
	if err != nil {
		t.Fatalf("Failed to read events cache file: %v", err)
	}
	if len(data) == 0 {
		t.Error("Events cache file is empty")
	}
}

func TestGetCachedEvents_NoCache(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Test with no cached events
	cachedEvents, err := getCachedEventsFromDir(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error getting cached events: %v", err)
	}

	if len(cachedEvents) != 0 {
		t.Errorf("Expected empty slice for no cache, got %d events", len(cachedEvents))
	}
}

func TestGetCachedEvents_CorruptCache(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	// Create corrupt cache file
	eventsFile := filepath.Join(tempDir, "david_events.json")
	err := os.WriteFile(eventsFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupt cache file: %v", err)
	}

	// Test with corrupt cache
	cachedEvents, err := getCachedEventsFromDir(tempDir)
	if err != nil {
		t.Fatalf("Unexpected error getting cached events: %v", err)
	}

	// Should return empty slice for corrupt cache
	if len(cachedEvents) != 0 {
		t.Errorf("Expected empty slice for corrupt cache, got %d events", len(cachedEvents))
	}
}

func TestCacheWrapperFunctions(t *testing.T) {
	// Test the wrapper functions that just call the directory-specific versions
	// These are currently at 0% coverage

	// Test getCachedETag wrapper
	etag := getCachedETag()
	// Should not panic, even if no cache exists
	_ = etag

	// Test cacheETag wrapper
	cacheETag("test-etag")
	// Should not panic

	// Test getCachedEvents wrapper
	events, err := getCachedEvents()
	// Should not panic
	if err != nil {
		t.Logf("getCachedEvents returned error (expected): %v", err)
	}
	_ = events

	// Test cacheEvents wrapper
	testEvents := []Event{
		{
			Code:    "1",
			StartAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			EndAt:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
		},
	}
	cacheEvents(testEvents)
	// Should not panic
}
