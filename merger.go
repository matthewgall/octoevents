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
	"sort"
	"strings"
	"time"
)

// hasChanges checks if there are any changes between existing and new events
func hasChanges(existing, new []Event) bool {
	if len(existing) != len(new) {
		return true
	}

	// Create a map of existing events by their unique key (start+end time)
	existingMap := make(map[string]bool)
	for _, event := range existing {
		key := event.StartAt.Format(time.RFC3339) + "_" + event.EndAt.Format(time.RFC3339)
		existingMap[key] = true
	}

	// Check if any new events are missing from existing
	for _, event := range new {
		key := event.StartAt.Format(time.RFC3339) + "_" + event.EndAt.Format(time.RFC3339)
		if !existingMap[key] {
			return true // Found a new event
		}
	}

	return false // No changes detected
}

// mergeEvents merges existing and new events, deduplicating by start+end time
func mergeEvents(existing, new []Event) []Event {
	// Pre-allocate map with estimated capacity
	capacity := len(existing) + len(new)
	eventMap := make(map[string]Event, capacity)

	for _, event := range existing {
		keyBuilder := builderPool.Get().(*strings.Builder)
		keyBuilder.Reset()
		keyBuilder.WriteString(event.StartAt.Format(time.RFC3339))
		keyBuilder.WriteByte('_')
		keyBuilder.WriteString(event.EndAt.Format(time.RFC3339))
		eventMap[keyBuilder.String()] = event
		builderPool.Put(keyBuilder)
	}

	for _, event := range new {
		keyBuilder := builderPool.Get().(*strings.Builder)
		keyBuilder.Reset()
		keyBuilder.WriteString(event.StartAt.Format(time.RFC3339))
		keyBuilder.WriteByte('_')
		keyBuilder.WriteString(event.EndAt.Format(time.RFC3339))
		eventMap[keyBuilder.String()] = event
		builderPool.Put(keyBuilder)
	}

	// Pre-allocate slice with exact capacity
	merged := make([]Event, 0, len(eventMap))
	for _, event := range eventMap {
		merged = append(merged, event)
	}

	sort.Slice(merged, func(i, j int) bool {
		return merged[i].StartAt.Before(merged[j].StartAt)
	})

	return merged
}
