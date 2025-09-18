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
	"encoding/json"
	"log/slog"
	"os"
)

const cacheDir = ".cache"

// getCachedETag retrieves the cached ETag for conditional requests
func getCachedETag() string {
	return getCachedETagFromDir(cacheDir)
}

// getCachedETagFromDir retrieves the cached ETag from a specific directory
func getCachedETagFromDir(cacheDir string) string {
	data, err := os.ReadFile(cacheDir + "/etag")
	if err != nil {
		slog.Debug("No cached ETag found", "error", err)
		return ""
	}
	etag := string(data)
	slog.Debug("Using cached ETag", "etag", etag)
	return etag
}

// cacheETag stores the ETag for future conditional requests
func cacheETag(etag string) {
	cacheETagToDir(cacheDir, etag)
}

// cacheETagToDir stores the ETag to a specific directory
func cacheETagToDir(cacheDir, etag string) {
	os.MkdirAll(cacheDir, 0755)
	if err := os.WriteFile(cacheDir+"/etag", []byte(etag), 0644); err != nil {
		slog.Warn("Failed to cache ETag", "error", err)
	} else {
		slog.Debug("Cached new ETag", "etag", etag)
	}
}

// getCachedEvents retrieves cached events from disk
func getCachedEvents() ([]Event, error) {
	return getCachedEventsFromDir(cacheDir)
}

// getCachedEventsFromDir retrieves cached events from a specific directory
func getCachedEventsFromDir(cacheDir string) ([]Event, error) {
	data, err := os.ReadFile(cacheDir + "/david_events.json")
	if err != nil {
		return []Event{}, nil // Return empty if no cache
	}

	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		return []Event{}, nil // Return empty if corrupt cache
	}

	return events, nil
}

// cacheEvents stores events to disk for future use
func cacheEvents(events []Event) {
	cacheEventsToDir(cacheDir, events)
}

// cacheEventsToDir stores events to a specific directory
func cacheEventsToDir(cacheDir string, events []Event) {
	os.MkdirAll(cacheDir, 0755)
	data, _ := json.Marshal(events)
	os.WriteFile(cacheDir+"/david_events.json", data, 0644)
}
