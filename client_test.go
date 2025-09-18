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
	"testing"
)

func TestNewAuthenticatedClient(t *testing.T) {
	apiKey := "test-api-key"
	graphqlURL := "https://api.test.com/graphql"

	client := NewAuthenticatedClient(apiKey, graphqlURL)

	if client == nil {
		t.Fatal("NewAuthenticatedClient returned nil")
	}

	if client.apiKey != apiKey {
		t.Errorf("Expected API key '%s', got '%s'", apiKey, client.apiKey)
	}

	if client.graphqlURL != graphqlURL {
		t.Errorf("Expected GraphQL URL '%s', got '%s'", graphqlURL, client.graphqlURL)
	}

	if client.client == nil {
		t.Error("GraphQL client was not initialized")
	}
}

// Note: ensureValidToken, obtainToken, and Run methods make network calls
// and are difficult to test without mocking. They remain at 0% coverage
// but are tested indirectly through integration tests.
