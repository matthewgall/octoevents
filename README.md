# OctoEvents

A Go-based service that fetches Octopus Energy free electricity events using their GraphQL API and maintains a continuously updated JSON file via GitHub Actions.

## Features

- Fetches free electricity events from Octopus Energy's GraphQL API
- Merges new events with existing ones (never deletes events)
- Prevents publishing empty files
- Runs automatically every hour via GitHub Actions
- Sorts events chronologically by start time

## Setup

### Configuration Methods

The application supports multiple configuration methods (in order of precedence):

1. **Command Line Arguments** (highest priority)
2. **Configuration File** (YAML)
3. **Environment Variables** (lowest priority)

### Command Line Usage

```bash
# Using command line flags
go run . -key sk_live_your_api_key_here -account A-12345678 -meter 1000000000000

# Using config file
go run . -config config.yaml

# Using environment variables
export OCTOPUS_API_KEY="sk_live_your_api_key_here"
export ACCOUNT_NUMBER="A-12345678" 
export METER_POINT_ID="1000000000000"
go run .

# Custom output file and logging
go run . -key sk_live_your_api_key_here -account A-12345678 -meter 1000000000000 -output events.json

# Log format options
go run . -config config.yaml -log-format text  # Human-readable logs
go run . -config config.yaml -log-format json  # Structured JSON logs  
go run . -config config.yaml -log-format auto  # Auto-detect (default)

# Show version
go run . -version
```

### Configuration File

Create a `config.yaml` file:

```yaml
accountNumber: A-12345678
meterPointID: "1000000000000"
apiKey: sk_live_your_api_key_here
outputFile: free_electricity.json
```

### GitHub Secrets

For GitHub Actions, configure these secrets:

- `OCTOPUS_API_KEY`: Your Octopus Energy API key (e.g., "sk_live_...")
- `ACCOUNT_NUMBER`: Your Octopus account number (e.g., "A-12345678")
- `METER_POINT_ID`: Your electricity meter point ID (MPAN, e.g., "1000000000000")

## Output

The service generates `free_electricity.json` containing an array of events with the following structure:

```json
{
  "data": [
    {
      "start": "2024-08-15T12:00:00.000Z",
      "end": "2024-08-15T13:00:00.000Z",
      "code": "1"
    },
    {
      "start": "2024-11-27T11:00:00.000Z",
      "end": "2024-11-27T12:00:00.000Z",
      "code": "2",
      "is_test": true
    }
  ]
}
```

### Fields

- **data**: Array wrapper containing all events
- **start**: Event start time in UTC (ISO 8601 format with milliseconds)
- **end**: Event end time in UTC (ISO 8601 format with milliseconds)  
- **code**: Sequential integer identifier starting from 1 (as string)
- **is_test**: Optional boolean flag indicating test events (only appears when true)

## How It Works

1. GitHub Actions runs the Go application every hour
2. The app authenticates with Octopus Energy using your API key to obtain a JWT token
3. Using the JWT token, it fetches current events from Octopus Energy's GraphQL API
4. Merges David Kendall's historical data with new events from Octopus GraphQL
5. Events are deduplicated using start+end time as unique identifiers
6. Sequential integer codes are assigned (1, 2, 3...)
7. The file is only updated if new events are found
8. Changes are automatically committed and deployed to GitHub Pages

This ensures a continuously growing dataset of historical and upcoming free electricity events.

## Public API

Once deployed, the data will be available at:
- **API Endpoint**: `https://matthewgall.github.io/octoevents/free_electricity.json`
- **Documentation**: `https://matthewgall.github.io/octoevents/`

The API provides:
- Real-time free electricity event data
- Historical events from David Kendall's API merged with new Octopus data  
- Sequential integer codes for easy reference
- Optional `is_test` flags for test events
- Automatic hourly updates
- High availability via GitHub Pages

## Authentication Flow

The service implements proper Octopus Energy authentication:
- Uses your API key to obtain a JWT token via GraphQL mutation
- Automatically refreshes the JWT token when it expires (with 5-minute buffer)
- Includes all required headers to match browser behavior
- Thread-safe token management with mutex locks

## GitHub Pages Setup

To enable GitHub Pages deployment:

1. Go to your repository Settings → Pages
2. Set Source to "GitHub Actions"  
3. The workflow will automatically deploy after each data update

## Building and Versioning

### Development Builds
```bash
go build .                    # Version: git commit hash or "dev"
go run . -version             # Show current version
```

### Release Builds
```bash
# Build with explicit version
go build -ldflags "-X main.buildVersion=1.0.0" -o octoevents .

# Build with version and commit
go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildCommit=$(git rev-parse HEAD)" .
```

### Version Detection Strategy
1. **Explicit version** via `-ldflags -X main.buildVersion=...` (highest priority)
2. **Git commit hash** from build info (auto-detected)
3. **Explicit commit** via `-ldflags -X main.buildCommit=...`
4. **Fallback** to "dev" for local development

## Performance Features

The service includes several optimizations:

- **HTTP Connection Pooling**: Reuses connections for better performance
- **Smart Caching**: Uses ETag conditional requests to avoid unnecessary downloads
- **GitHub Actions Cache**: Persists cache data across workflow runs (daily rotation)
- **Memory Optimization**: Pre-allocated data structures and efficient string handling
- **Structured Logging**: JSON-formatted logs with contextual information
- **Graceful Fallback**: Continues operation even if external API is unavailable

Cache benefits:
- **First run**: Downloads David Kendall's data (~18 events)
- **Subsequent runs**: Uses 304 Not Modified responses when data unchanged
- **Daily rotation**: Cache keys rotate daily to ensure freshness

## Support the Project

If you find OctoJoin useful, here are some ways to support its continued development:

### 💷 Join Octopus Energy

Not an Octopus Energy customer yet? Use my referral link to join and we'll both get £50 credit:

**[Join Octopus Energy - Get £50 credit](https://share.octopus.energy/maize-ape-570)**

This helps fund development of OctoJoin and you'll get access to:
- Saving Sessions (earn money for reducing usage during peak times)
- Free electricity sessions (completely free electricity during certain periods)
- Competitive energy rates and excellent customer service
- The greenest energy supplier in the UK

### ❤️ GitHub Sponsor

Support ongoing development and maintenance:

**[Become a GitHub Sponsor](https://github.com/sponsors/matthewgall)**

Your sponsorship helps with:
- Adding new features and improvements
- Maintaining compatibility with API changes  
- Providing support and bug fixes
- Keeping the project free and open source

### ⭐ Star the Repository

Show your appreciation by starring the repository on GitHub - it helps others discover the project!

## Disclaimer

This is an unofficial third-party application developed independently. "Octopus Energy" is a trademark of Octopus Energy Group Limited. This application is not affiliated with, endorsed by, or connected to Octopus Energy Group Limited in any way.

The application interacts with publicly documented Octopus Energy APIs for data gathering purposes. Users are responsible for ensuring their use complies with Octopus Energy's terms of service.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.