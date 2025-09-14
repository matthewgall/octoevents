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
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type Event struct {
	Code               string    `json:"code"`
	EndAt              time.Time `json:"endAt"`
	IsEventParticipant bool      `json:"isEventParticipant"`
	Name               string    `json:"name"`
	StartAt            time.Time `json:"startAt"`
	Typename           string    `json:"__typename"`
	IsTest             *bool     `json:"isTest,omitempty"`
}

type OutputEvent struct {
	Start  string `json:"start"`
	End    string `json:"end"`
	Code   string `json:"code"`
	IsTest *bool  `json:"is_test,omitempty"`
}

type OutputData struct {
	Data []OutputEvent `json:"data"`
}

type EventEdge struct {
	Cursor string `json:"cursor"`
	Node   Event  `json:"node"`
}

type PageInfo struct {
	EndCursor       string `json:"endCursor"`
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
}

type EventConnection struct {
	Edges      []EventEdge `json:"edges"`
	PageInfo   PageInfo    `json:"pageInfo"`
	TotalCount int         `json:"totalCount"`
	EdgeCount  int         `json:"edgeCount"`
}

type GraphQLResponse struct {
	IsEnrolledInCustomerFlexibilityCampaign    bool            `json:"isEnrolledInCustomerFlexibilityCampaign"`
	CustomerFlexibilityCampaignEvents          EventConnection `json:"customerFlexibilityCampaignEvents"`
}

const (
	graphqlEndpoint = "https://api.octopus.energy/v1/graphql/"
	davidKendallAPI = "https://oe-api.davidskendall.co.uk/free_electricity.json"
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

func setupLogging() {
	var handler slog.Handler
	
	format := *logFormat
	if format == "auto" {
		format = detectLogFormat()
	}

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	switch format {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		// Default to text for unknown formats
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func detectLogFormat() string {
	// Use JSON format in GitHub Actions or other CI environments
	if os.Getenv("GITHUB_ACTIONS") == "true" || 
	   os.Getenv("CI") == "true" ||
	   os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return "json"
	}
	
	// Use text format for local development
	return "text"
}


func fetchAndUpdateEvents(config *Config) error {
	client := NewAuthenticatedClient(config.APIKey, graphqlEndpoint)

	query := `
		query getFreeElectricityEnrollmentAndEvents($accountNumber: String!, $meterPointId: String!, $campaignSlug: String!) {
			isEnrolledInCustomerFlexibilityCampaign(
				accountNumber: $accountNumber
				campaignSlug: $campaignSlug
				supplyPointIdentifier: $meterPointId
			)
			customerFlexibilityCampaignEvents(
				accountNumber: $accountNumber
				campaignSlug: $campaignSlug
				supplyPointIdentifier: $meterPointId
				first: 20
			) {
				edges {
					cursor
					node {
						code
						endAt
						isEventParticipant
						name
						startAt
						__typename
					}
					__typename
				}
				pageInfo {
					endCursor
					hasNextPage
					hasPreviousPage
					startCursor
					__typename
				}
				totalCount
				edgeCount
				__typename
			}
		}
	`

	req := graphql.NewRequest(query)
	req.Var("accountNumber", config.AccountNumber)
	req.Var("meterPointId", config.MeterPointID)
	req.Var("campaignSlug", "free_electricity")


	var response GraphQLResponse
	if err := client.Run(context.Background(), req, &response); err != nil {
		return errors.Wrap(err, "failed to execute GraphQL query")
	}

	newEvents := make([]Event, 0, len(response.CustomerFlexibilityCampaignEvents.Edges))
	for _, edge := range response.CustomerFlexibilityCampaignEvents.Edges {
		newEvents = append(newEvents, edge.Node)
	}

	if len(newEvents) == 0 {
		return fmt.Errorf("no events received from API - refusing to update file")
	}

	existingEvents, err := loadExistingEvents(config.OutputFile)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to load existing events")
	}

	// Fetch David Kendall's data to merge
	externalEvents, err := fetchDavidKendallData()
	if err != nil {
		slog.Warn("Failed to fetch David Kendall's data", "error", err)
		externalEvents = []Event{}
	}

	// Merge all events
	allExistingEvents := mergeEvents(existingEvents, externalEvents)
	mergedEvents := mergeEvents(allExistingEvents, newEvents)

	// Assign sequential codes
	mergedEvents = assignSequentialCodes(mergedEvents)

	if err := saveEvents(mergedEvents, config.OutputFile); err != nil {
		return errors.Wrap(err, "failed to save events")
	}

	slog.Info("Successfully updated events", 
		"file", config.OutputFile, 
		"count", len(mergedEvents),
		"new_events", len(newEvents),
		"external_events", len(externalEvents))
	return nil
}

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

func mergeEvents(existing, new []Event) []Event {
	// Pre-allocate map with estimated capacity
	capacity := len(existing) + len(new)
	eventMap := make(map[string]Event, capacity)

	// Use string builder for more efficient key generation
	var keyBuilder strings.Builder
	keyBuilder.Grow(64) // Pre-allocate for typical RFC3339 timestamp pairs

	for _, event := range existing {
		keyBuilder.Reset()
		keyBuilder.WriteString(event.StartAt.Format(time.RFC3339))
		keyBuilder.WriteByte('_')
		keyBuilder.WriteString(event.EndAt.Format(time.RFC3339))
		eventMap[keyBuilder.String()] = event
	}

	for _, event := range new {
		keyBuilder.Reset()
		keyBuilder.WriteString(event.StartAt.Format(time.RFC3339))
		keyBuilder.WriteByte('_')
		keyBuilder.WriteString(event.EndAt.Format(time.RFC3339))
		eventMap[keyBuilder.String()] = event
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

func fetchDavidKendallData() ([]Event, error) {
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}

	req, err := http.NewRequest("GET", davidKendallAPI, nil)
	if err != nil {
		return nil, err
	}

	// Add conditional request headers for caching
	req.Header.Set("User-Agent", GetUserAgent())
	req.Header.Set("Accept", "application/json")
	
	// Check if we have cached ETag
	if etag := getCachedETag(); etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Handle 304 Not Modified
	if resp.StatusCode == http.StatusNotModified {
		slog.Info("David Kendall's API data unchanged", "status", 304)
		return getCachedEvents()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var outputData OutputData
	if err := json.NewDecoder(resp.Body).Decode(&outputData); err != nil {
		return nil, err
	}

	// Cache the ETag for next request
	if etag := resp.Header.Get("ETag"); etag != "" {
		cacheETag(etag)
	}

	// Convert to internal format
	events := make([]Event, 0, len(outputData.Data))
	for _, outputEvent := range outputData.Data {
		startTime, err := time.Parse("2006-01-02T15:04:05.000Z", outputEvent.Start)
		if err != nil {
			continue // Skip invalid entries
		}
		endTime, err := time.Parse("2006-01-02T15:04:05.000Z", outputEvent.End)
		if err != nil {
			continue // Skip invalid entries
		}

		event := Event{
			Code:    outputEvent.Code,
			StartAt: startTime,
			EndAt:   endTime,
			IsTest:  outputEvent.IsTest,
		}
		events = append(events, event)
	}

	// Cache the events
	cacheEvents(events)
	
	slog.Info("Fetched events from David Kendall's API", "count", len(events))
	return events, nil
}

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

const cacheDir = ".cache"

func getCachedETag() string {
	data, err := os.ReadFile(cacheDir + "/etag")
	if err != nil {
		slog.Debug("No cached ETag found", "error", err)
		return ""
	}
	etag := string(data)
	slog.Debug("Using cached ETag", "etag", etag)
	return etag
}

func cacheETag(etag string) {
	os.MkdirAll(cacheDir, 0755)
	if err := os.WriteFile(cacheDir+"/etag", []byte(etag), 0644); err != nil {
		slog.Warn("Failed to cache ETag", "error", err)
	} else {
		slog.Debug("Cached new ETag", "etag", etag)
	}
}

func getCachedEvents() ([]Event, error) {
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

func cacheEvents(events []Event) {
	os.MkdirAll(cacheDir, 0755)
	data, _ := json.Marshal(events)
	os.WriteFile(cacheDir+"/david_events.json", data, 0644)
}

func saveEvents(events []Event, filename string) error {
	outputData := convertToOutputFormat(events)
	data, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}