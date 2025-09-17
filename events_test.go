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
	"reflect"
	"testing"
	"time"
)

func TestAssignSequentialCodes(t *testing.T) {
	events := []Event{
		{StartAt: time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)},
		{StartAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{StartAt: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
	}

	result := assignSequentialCodes(events)

	// Should be sorted by start time and assigned sequential codes
	expected := []string{"1", "2", "3"}
	actual := []string{result[0].Code, result[1].Code, result[2].Code}

	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected codes %v, got %v", expected, actual)
	}

	// Verify events are sorted by start time
	if !result[0].StartAt.Before(result[1].StartAt) || !result[1].StartAt.Before(result[2].StartAt) {
		t.Error("Events should be sorted by start time")
	}
}

func TestConvertToOutputFormat(t *testing.T) {
	events := []Event{
		{
			Code:    "1",
			StartAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			EndAt:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			IsTest:  boolPtr(true),
		},
	}

	result := convertToOutputFormat(events)

	if len(result.Data) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(result.Data))
	}

	expected := OutputEvent{
		Start:  "2024-01-01T12:00:00.000Z",
		End:    "2024-01-01T13:00:00.000Z",
		Code:   "1",
		IsTest: boolPtr(true),
	}

	if !reflect.DeepEqual(result.Data[0], expected) {
		t.Errorf("Expected %+v, got %+v", expected, result.Data[0])
	}
}

func TestLoadExistingEvents(t *testing.T) {
	// Create a temporary file with test data
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.json")

	testData := `{
		"data": [
			{
				"start": "2024-01-01T12:00:00.000Z",
				"end": "2024-01-01T13:00:00.000Z",
				"code": "1",
				"is_test": true
			}
		]
	}`

	err := os.WriteFile(testFile, []byte(testData), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	events, err := loadExistingEvents(testFile)
	if err != nil {
		t.Fatalf("Failed to load events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	expectedStart, _ := time.Parse("2006-01-02T15:04:05.000Z", "2024-01-01T12:00:00.000Z")
	expectedEnd, _ := time.Parse("2006-01-02T15:04:05.000Z", "2024-01-01T13:00:00.000Z")

	if events[0].Code != "1" || !events[0].StartAt.Equal(expectedStart) || !events[0].EndAt.Equal(expectedEnd) {
		t.Errorf("Event data doesn't match expected values")
	}
}

func TestHasChanges(t *testing.T) {
	// Test with different lengths
	existing := []Event{{}}
	new := []Event{{}, {}}
	if !hasChanges(existing, new) {
		t.Error("Expected changes when lengths differ")
	}

	// Test with same data
	event1 := Event{StartAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), EndAt: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)}
	event2 := Event{StartAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), EndAt: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)}
	if hasChanges([]Event{event1}, []Event{event2}) {
		t.Error("Expected no changes for identical events")
	}

	// Test with different data
	event3 := Event{StartAt: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC), EndAt: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)}
	if !hasChanges([]Event{event1}, []Event{event3}) {
		t.Error("Expected changes for different events")
	}
}

func TestMergeEvents(t *testing.T) {
	event1 := Event{StartAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), EndAt: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)}
	event2 := Event{StartAt: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC), EndAt: time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)}
	event3 := Event{StartAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC), EndAt: time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)} // Duplicate

	result := mergeEvents([]Event{event1, event2}, []Event{event3})

	// Should have 2 unique events
	if len(result) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(result))
	}

	// Should be sorted by start time
	if !result[0].StartAt.Before(result[1].StartAt) {
		t.Error("Events should be sorted by start time")
	}
}

func TestSaveEvents(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "output.json")

	events := []Event{
		{
			Code:    "1",
			StartAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			EndAt:   time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC),
			IsTest:  boolPtr(true),
		},
	}

	err := saveEvents(events, testFile)
	if err != nil {
		t.Fatalf("Failed to save events: %v", err)
	}

	// Verify file was created and has content
	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Saved file is empty")
	}

	// Verify it can be loaded back
	loadedEvents, err := loadExistingEvents(testFile)
	if err != nil {
		t.Fatalf("Failed to load saved events: %v", err)
	}

	if len(loadedEvents) != 1 {
		t.Fatalf("Expected 1 loaded event, got %d", len(loadedEvents))
	}
}

// Helper function to create bool pointer
func boolPtr(b bool) *bool {
	return &b
}
