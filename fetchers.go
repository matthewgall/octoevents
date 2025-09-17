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
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// fetchOctopusEvents fetches events from the Octopus Energy GraphQL API
func fetchOctopusEvents(config *Config) ([]Event, error) {
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
		return nil, errors.Wrap(err, "failed to execute GraphQL query")
	}

	events := make([]Event, 0, len(response.CustomerFlexibilityCampaignEvents.Edges))
	for _, edge := range response.CustomerFlexibilityCampaignEvents.Edges {
		events = append(events, edge.Node)
	}

	return events, nil
}

// fetchDavidKendallData fetches events from David Kendall's API with caching
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
