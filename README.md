# User Preferences Service

A centralized service for managing user preferences across SoundCloud.

## Overview

The User Preferences Service provides a standardized way to manage user preferences, such as:

- Badge visibility
- Featured fans settings
- Notification preferences
- Privacy settings

### Entry Point (cmd/api/main.go)

This file serves as the application bootstrap, initializing the application context and wiring up
all dependencies. It creates the Twirp server for handling user settings requests and registers
the router with the HTTP server. The main function orchestrates the initialization sequence:
setting up the context with DynamoDB client, repository, and service, then starting the HTTP
server to listen for incoming requests.

### API Handlers (internal/api/settings.go)

The API layer provides three main endpoints: GetSettings, PutSettings, and DeleteSettings.
Each handler performs input validation, calls the appropriate service method, and transforms
the result into the response format. The handlers act as a translation layer between the
JSON/Protocol Buffer requests and the internal domain models. Error handling is comprehensive,
with appropriate error codes returned for validation failures, not found conditions, and server errors.

### Service Layer (internal/prefsservice/service.go)

The service implementation encapsulates the business logic for user settings operations.
It validates that setting values contain valid JSON, handles type conversions between internal
models and protocol buffer types, and manages the setting types system. The service layer
defines interfaces that decouple implementation details, making the system more testable
and maintainable. It acts as the intermediary between the API layer and data access layer.

### Minor protection User Attributes

Service is calling Okidoki to fetch user data and determine if the user is minor. 
Decision logic is located in /regionalrulesrepo/repository.go
RulesService fetching rules set for the particular user and based on the rules provides RulesApplication value.

```go
const (
	ApplyRestriction RuleApplication = iota
	NoRestriction
)
```

```go
rules := []Rule{{
        ID:    "GB",
        Type:  RuleAgeRestriction,
        Value: "{\"age\": 18}",
    }, {
        ID:    "US",
        Type:  RuleAgeRestriction,
        Value: "{\"age\": 18}",
    }}
```

### Service extension in the future

In the future we could extend setting with the metadata, which could be stored separate from the settings themselves.
We could include following info as a definitions reference and utilize it on a setting creation:

| Field    | Type    | Notes                                                              |
|----------|---------|--------------------------------------------------------------------|
| type     | string  | S or B or N                                                        |
| privacy  | boolean | Appears in GetPrivacySettings response, by other users, if enabled |
| default  | string  |                                                                    |                                                                    |
| category | string  |                                                                    |                                                                    |
| order    | number  |                                                                    |                                                                    |

### Repository Layer (internal/prefsrepo/repository.go)

This repository provides a clean abstraction over DynamoDB operations. It implements methods
for creating, updating, retrieving, and deleting user settings. The repository uses DynamoDB
expressions to build queries and conditions for data access. It handles marshaling/unmarshaling
between Go structs and DynamoDB attribute values. Primary and sort keys are constructed to
enable efficient lookups by user URN and setting Name.

### Data Models (internal/prefsrepo/models.go)

The models define structure of user settings as they're stored in the database. Each setting
has properties like Name, user URN, type, updated timestamp, and value.
The models include methods for converting between internal representations and database items.
The setting types are represented as strings internally but mapped to enum values in the API.

### Application Context (internal/appserver/context.go)

The context provides dependency injection for the app, creating and maintaining references to
the repository and service implementations. It inits DynamoDB client with config and sets up
OpenTelemetry for observability. The context ensures that dependencies are properly initialized
before the server starts, using a singleton pattern with a sync.Once to prevent multiple instances.

### Server Configuration (internal/appserver/server.go)

This component handles HTTP server lifecycle, including graceful shutdown on sigint. It configures
the server with the router and address, starts listening for requests, and waits for termination signals.

## Protocol Buffer Schema for User Settings

Location: rpc/user_settings_api.proto
Interface definition for client-server comms via Twirp RPC

### Setting Types Enum

```proto
enum SettingType {
  STRING = 0;
  BOOLEAN = 1;
  NUMBER = 2;
}
```

### Design

- Strong typing for the setting Value
- Service layer handles conversions between proto/internal formats

```proto
service UserSettingsService {
  rpc GetSettings(GetSettingsRequest) returns (GetSettingsResponse) {}
  rpc GetPrivacySettings(GetPrivacySettingsRequest) returns (GetPrivacySettingsResponse) {}
  rpc PutSetting(PutSettingRequest) returns (PutSettingResponse) {}

  rpc FetchUsersPrivacyAttributes(UsersPrivacyAttributesRequest) returns (UsersPrivacyAttributesResponse) {}
}

```

# Database Schema

## User Settings Table

| Field       | Type      | Index | Notes |
|-------------|-----------|-------|-------|
| pk(userUrn) | string    | Hash  |       |
| sk(name)    | string    | Range |       |
| type        | string    |       | S/B/N |
| value       | string    |       |       |
| updatedAt   | time.Time |       |       |

## Prerequisites

- [Go 1.21+](https://go.dev/doc/install)
- [Protobuf Compiler](https://grpc.io/docs/protoc-installation/)
- [Task](https://taskfile.dev/installation/) (task runner)

## Getting Started

1. Install dependencies:

   ```bash
   task deps:macos
   task install
   ```

2. Generate protocol buffer code:

   ```bash
   task gen:rpc
   ```

3. Build the service:

   ```bash
   task build
   ```

4. Run the service:

   ```bash
   task dev
   ```

5. Test the service:

   ```bash
   task test
   ```

## API Endpoints

The service provides a [Twirp](https://github.com/twitchtv/twirp) RPC API:

```text
/twirp/proto.soundcloud.user_settings.api.UserSettingsService/*
```

### Available Endpoints

- **GetSettings**: Get preferences for a user based on UserSession
- **GetPrivacySettings**: Gets privacy related preferences for a User by UserUrn
- **PutSetting**: Create or update a preference
- **FetchUsersPrivacyAttributes**: Provides user privacy attributes based on UserUrn

## Sample Requests

### Get User Preferences

Use this endpoint to retrieve preferences for a given user. In the request, provide your session
details along with the user URN and at least one setting ID you wish to fetch.

```bash
curl -i -X POST -H "Content-Type: application/json" \
  http://localhost:8000/twirp/proto.soundcloud.user_settings.api.UserSettingsService/GetSettings \
  --data '{
    "user_session": {
      "user_urn": "soundcloud:users:123456",
      "agent_urn": "urn:agent:abc",
      "geo": {
        "country_code": "US"
      },
      "scopes": [],
      "features": []
    },
    "setting_names": ["setting1"]
  }'
```

```bash
curl -i -X POST -H "Content-Type: application/json" \
  http://localhost:8000/twirp/proto.soundcloud.user_settings.api.UserSettingsService/GetPrivacySettings \
  --data '{
    "user_session": {
      "user_urn": "soundcloud:users:123456",
      "agent_urn": "urn:agent:abc",
      "geo": {
        "country_code": "US"
      },
      "scopes": [],
      "features": []
    },
    "user_urns": ["soundcloud:users:654321"]
  }'
```

### Put a User Preference

This endpoint creates a new preference or updates an existing one. In the sample below, a badge
preference is being Puted. Note that the service will update the `updated_at` timestamp automatically.

```bash
curl -i -X POST -H "Content-Type: application/json" \
  http://localhost:8000/twirp/proto.soundcloud.user_settings.api.UserSettingsService/PutSetting \
  --data '{
    "user_session": {
      "user_urn": "soundcloud:users:123456",
      "agent_urn": "urn:agent:abc",
      "geo": {
        "country_code": "US"
      },
      "scopes": [],
      "features": []
    },
    "setting": {
      "name": "setting1",
      "type": 1,
      "value": "true"
    }
  }'
```

### Get User Privacy Attributes

Use this endpoint to fetch privacy attributes for a set of Users. Provide your session data along with
the user URNs.

```bash
curl -i -X POST -H "Content-Type: application/json" \
  http://localhost:8000/twirp/proto.soundcloud.user_settings.api.UserSettingsService/FetchUsersPrivacyAttributes \
  --data '{
    "user_session": {
      "user_urn": "soundcloud:users:123456",
      "agent_urn": "urn:agent:abc",
      "geo": {
        "country_code": "US"
      },
      "scopes": [],
      "features": []
    },
    "user_urns": ["soundcloud:users:788871454"]
  }'
```

Thank you @bbeesley and @caplis for your foundational work that helped me piece this together!
### Resources:
[https://github.com/soundcloud/fortune-go](https://github.com/soundcloud/fortune-go)
[https://github.com/soundcloud/private-resource-authorizer](https://github.com/soundcloud/private-resource-authorizer)
