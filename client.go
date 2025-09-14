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
	"net/http"
	"sync"
	"time"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

type AuthenticatedClient struct {
	apiKey       string
	graphqlURL   string
	client       *graphql.Client
	token        string
	tokenExpiry  time.Time
	refreshToken string
	mutex        sync.RWMutex
}

type ObtainTokenInput struct {
	APIKey string `json:"APIKey"`
}

type TokenResponse struct {
	Token            string `json:"token"`
	RefreshToken     string `json:"refreshToken"`
	RefreshExpiresIn int    `json:"refreshExpiresIn"`
}

type ObtainTokenMutation struct {
	ObtainKrakenToken TokenResponse `json:"obtainKrakenToken"`
}

func NewAuthenticatedClient(apiKey, graphqlURL string) *AuthenticatedClient {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
		},
	}

	client := graphql.NewClient(graphqlURL, graphql.WithHTTPClient(httpClient))

	return &AuthenticatedClient{
		apiKey:     apiKey,
		graphqlURL: graphqlURL,
		client:     client,
	}
}

func (c *AuthenticatedClient) ensureValidToken(ctx context.Context) error {
	c.mutex.RLock()
	hasValidToken := c.token != "" && time.Now().Add(5*time.Minute).Before(c.tokenExpiry)
	c.mutex.RUnlock()

	if hasValidToken {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.token != "" && time.Now().Add(5*time.Minute).Before(c.tokenExpiry) {
		return nil
	}

	return c.obtainToken(ctx)
}

func (c *AuthenticatedClient) obtainToken(ctx context.Context) error {
	mutation := `
		mutation obtainKrakenToken($input: ObtainJSONWebTokenInput!) {
			obtainKrakenToken(input: $input) {
				token
				refreshToken
				refreshExpiresIn
			}
		}
	`

	req := graphql.NewRequest(mutation)
	req.Var("input", ObtainTokenInput{
		APIKey: c.apiKey,
	})

	req.Header.Set("Content-Type", "application/json")

	var response ObtainTokenMutation
	if err := c.client.Run(ctx, req, &response); err != nil {
		return errors.Wrap(err, "failed to obtain JWT token")
	}

	c.token = response.ObtainKrakenToken.Token
	c.refreshToken = response.ObtainKrakenToken.RefreshToken
	c.tokenExpiry = time.Now().Add(time.Duration(response.ObtainKrakenToken.RefreshExpiresIn) * time.Second)

	return nil
}

func (c *AuthenticatedClient) Run(ctx context.Context, req *graphql.Request, resp interface{}) error {
	if err := c.ensureValidToken(ctx); err != nil {
		return errors.Wrap(err, "failed to ensure valid token")
	}

	c.mutex.RLock()
	token := c.token
	c.mutex.RUnlock()

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	return c.client.Run(ctx, req, resp)
}
