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
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Event represents an electricity event
type Event struct {
	Code               string    `json:"code"`
	EndAt              time.Time `json:"endAt"`
	IsEventParticipant bool      `json:"isEventParticipant"`
	Name               string    `json:"name"`
	StartAt            time.Time `json:"startAt"`
	Typename           string    `json:"__typename"`
	IsTest             *bool     `json:"isTest,omitempty"`
}

// OutputEvent represents the output format for events
type OutputEvent struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Code   string `json:"code"`
	IsTest *bool  `json:"is_test,omitempty"`
}

// OutputData represents the complete output structure
type OutputData struct {
	Data []OutputEvent `json:"data"`
}

// EventEdge represents a GraphQL edge containing an event
type EventEdge struct {
	Cursor string `json:"cursor"`
	Node   Event  `json:"node"`
}

// PageInfo represents GraphQL pagination information
type PageInfo struct {
	EndCursor       string `json:"endCursor"`
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
}

// EventConnection represents a GraphQL connection of events
type EventConnection struct {
	Edges      []EventEdge `json:"edges"`
	PageInfo   PageInfo    `json:"pageInfo"`
	TotalCount int         `json:"totalCount"`
	EdgeCount  int         `json:"edgeCount"`
}

// GraphQLResponse represents the response from Octopus GraphQL API
type GraphQLResponse struct {
	IsEnrolledInCustomerFlexibilityCampaign bool            `json:"isEnrolledInCustomerFlexibilityCampaign"`
	CustomerFlexibilityCampaignEvents       EventConnection `json:"customerFlexibilityCampaignEvents"`
}

// builderPool provides a pool of string builders for efficient memory usage
var builderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

// assignSequentialCodes assigns sequential codes to events starting from 1
func assignSequentialCodes(events []Event) []Event {
	// Sort by start time to ensure consistent ordering
	sort.Slice(events, func(i, j int) bool {
		return events[i].StartAt.Before(events[j].StartAt)
	})

	// Assign sequential codes starting from 1
	for i := range events {
		events[i].Code = strconv.Itoa(i + 1)
	}

	return events
}

// convertToOutputFormat converts internal Event format to OutputData format
func convertToOutputFormat(events []Event) OutputData {
	outputEvents := make([]OutputEvent, 0, len(events))

	for _, event := range events {
		outputEvent := OutputEvent{
			Start:  event.StartAt.Format("2006-01-02T15:04:05.000Z"),
			End:    event.EndAt.Format("2006-01-02T15:04:05.000Z"),
			Code:   event.Code,
			IsTest: event.IsTest,
		}
		outputEvents = append(outputEvents, outputEvent)
	}

	return OutputData{Data: outputEvents}
}

// loadExistingEvents loads events from the output file
func loadExistingEvents(filename string) ([]Event, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var outputData OutputData
	if err := json.Unmarshal(data, &outputData); err != nil {
		return nil, err
	}

	// Convert back to internal format
	events := make([]Event, 0, len(outputData.Data))
	for _, outputEvent := range outputData.Data {
		startTime, err := time.Parse("2006-01-02T15:04:05.000Z", outputEvent.Start)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}
		endTime, err := time.Parse("2006-01-02T15:04:05.000Z", outputEvent.End)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}

		event := Event{
			Code:    outputEvent.Code,
			StartAt: startTime,
			EndAt:   endTime,
			IsTest:  outputEvent.IsTest,
		}
		events = append(events, event)
	}
	return events, nil
}

// saveEvents saves events to the output file
func saveEvents(events []Event, filename string) error {
	outputData := convertToOutputFormat(events)
	data, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
