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
	"log/slog"
	"os"

	"github.com/pkg/errors"
)

func main() {
	// Parse flags first to get log format preference
	flag.Parse()

	// Setup logging based on format preference
	setupLogging()

	config, err := loadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("Starting octoevents", "version", GetVersion())

	if err := fetchAndUpdateEvents(config); err != nil {
		slog.Error("Failed to fetch and update events", "error", err)
		os.Exit(1)
	}

	slog.Info("Successfully completed event update")
}

func fetchAndUpdateEvents(config *Config) error {
	// Always load existing events first - this is our safety net
	existingEvents, err := loadExistingEvents(config.OutputFile)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to load existing events")
	}

	slog.Info("Loaded existing events", "count", len(existingEvents))

	// Fetch events from both APIs concurrently
	type fetchResult struct {
		events []Event
		source string
		err    error
	}

	results := make(chan fetchResult, 2)

	// Fetch Octopus events
	go func() {
		events, err := fetchOctopusEvents(config)
		results <- fetchResult{events: events, source: "octopus", err: err}
	}()

	// Fetch David Kendall's data
	go func() {
		events, err := fetchDavidKendallData()
		results <- fetchResult{events: events, source: "david_kendall", err: err}
	}()

	// Collect results
	var octopusEvents, externalEvents []Event
	for i := 0; i < 2; i++ {
		result := <-results
		if result.err != nil {
			slog.Warn("Failed to fetch events", "source", result.source, "error", result.err)
			if result.source == "octopus" {
				octopusEvents = []Event{}
			} else {
				externalEvents = []Event{}
			}
		} else {
			slog.Info("Fetched events", "source", result.source, "count", len(result.events))
			if result.source == "octopus" {
				octopusEvents = result.events
			} else {
				externalEvents = result.events
			}
		}
	}

	// Start with existing events as the base (never lose data)
	allEvents := make([]Event, len(existingEvents))
	copy(allEvents, existingEvents)

	// Merge in external events if we got any
	if len(externalEvents) > 0 {
		allEvents = mergeEvents(allEvents, externalEvents)
	}

	// Merge in new Octopus events if we got any
	if len(octopusEvents) > 0 {
		allEvents = mergeEvents(allEvents, octopusEvents)
	}

	// Check if we actually have any changes
	if !hasChanges(existingEvents, allEvents) {
		slog.Info("No new events detected, skipping file update")
		return nil
	}

	// Assign sequential codes to the final merged set
	finalEvents := assignSequentialCodes(allEvents)

	// Final safety check: never write fewer events than we started with
	if len(finalEvents) < len(existingEvents) {
		slog.Warn("Refusing to write fewer events than existing",
			"existing", len(existingEvents),
			"new", len(finalEvents))
		return fmt.Errorf("safety check failed: would reduce event count from %d to %d",
			len(existingEvents), len(finalEvents))
	}

	// Save the updated events
	if err := saveEvents(finalEvents, config.OutputFile); err != nil {
		return errors.Wrap(err, "failed to save events")
	}

	slog.Info("Successfully updated events",
		"file", config.OutputFile,
		"total_count", len(finalEvents),
		"existing_count", len(existingEvents),
		"octopus_events", len(octopusEvents),
		"external_events", len(externalEvents),
		"new_events_added", len(finalEvents)-len(existingEvents))

	return nil
}
