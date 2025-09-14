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
	"runtime/debug"
)

// These variables are set at build time using ldflags
var (
	buildVersion = "dev"
	buildCommit  = "unknown"
)

// GetVersion returns the application version using a fallback strategy
func GetVersion() string {
	// If version is explicitly set via ldflags, use it
	if buildVersion != "dev" && buildVersion != "" {
		return buildVersion
	}

	// Try to get version from build info
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		// Look for version in build settings
		for _, setting := range buildInfo.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				// Use short commit hash (7 characters)
				if len(setting.Value) >= 7 {
					return setting.Value[:7]
				}
				return setting.Value
			}
		}
	}

	// If commit was set via ldflags, use it
	if buildCommit != "unknown" && buildCommit != "" {
		// Use short commit hash
		if len(buildCommit) >= 7 {
			return buildCommit[:7]
		}
		return buildCommit
	}

	// Fallback to "dev"
	return "dev"
}

// GetUserAgent returns a user agent string for HTTP requests
func GetUserAgent() string {
	return "matthewgall/octoevents/" + GetVersion()
}