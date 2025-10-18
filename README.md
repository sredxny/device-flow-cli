# device-flow-cli

A CLI tool that authenticates users via browser-based OAuth flow.

## Features

- Browser-based OAuth 2.0 authentication flow
- Automatic browser opening for seamless login experience
- Secure token storage in user's home directory
- Support for multiple OAuth providers (Google, GitHub, Auth0, etc.)
- CSRF protection with state parameter
- Token expiry checking

## Installation

```bash
go install github.com/sredxny/device-flow-cli@latest
```

Or build from source:

```bash
git clone https://github.com/sredxny/device-flow-cli.git
cd device-flow-cli
go build -o device-flow-cli
```

## Configuration

### Using Environment Variables

Copy `.env.example` to `.env` and fill in your OAuth provider details:

```bash
cp .env.example .env
```

Then edit `.env` with your OAuth configuration:

```
OAUTH_CLIENT_ID=your-client-id
OAUTH_CLIENT_SECRET=your-client-secret
OAUTH_AUTH_URL=https://accounts.google.com/o/oauth2/v2/auth
OAUTH_TOKEN_URL=https://oauth2.googleapis.com/token
```

### Using Command-Line Flags

You can also provide configuration via command-line flags:

```bash
device-flow-cli login \
  --client-id YOUR_CLIENT_ID \
  --client-secret YOUR_CLIENT_SECRET \
  --auth-url https://accounts.google.com/o/oauth2/v2/auth \
  --token-url https://oauth2.googleapis.com/token
```

## Usage

### Login

Authenticate by opening a browser window:

```bash
device-flow-cli login
```

This will:
1. Start a local callback server on port 8080
2. Open your default browser to the OAuth provider's login page
3. Wait for you to complete authentication
4. Receive the callback and store the access token securely

### Check Status

Check if you're currently authenticated:

```bash
device-flow-cli status
```

### Logout

Clear stored credentials:

```bash
device-flow-cli logout
```

## OAuth Provider Setup

### Registering Your Application

You need to register your application with an OAuth provider. Here are common examples:

#### Google OAuth

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project or select existing one
3. Enable the relevant APIs
4. Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client ID"
5. Select "Desktop app" as application type
6. Add `http://localhost:8080/callback` as authorized redirect URI

Configuration:
```
OAUTH_AUTH_URL=https://accounts.google.com/o/oauth2/v2/auth
OAUTH_TOKEN_URL=https://oauth2.googleapis.com/token
```

#### GitHub OAuth

1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Click "New OAuth App"
3. Set Homepage URL to your repo
4. Set Authorization callback URL to `http://localhost:8080/callback`

Configuration:
```
OAUTH_AUTH_URL=https://github.com/login/oauth/authorize
OAUTH_TOKEN_URL=https://github.com/login/oauth/access_token
```

#### Auth0

1. Go to your Auth0 Dashboard
2. Create a new Application (Native type)
3. Add `http://localhost:8080/callback` to Allowed Callback URLs

Configuration:
```
OAUTH_AUTH_URL=https://YOUR_DOMAIN.auth0.com/authorize
OAUTH_TOKEN_URL=https://YOUR_DOMAIN.auth0.com/oauth/token
```

## Security

- Tokens are stored in `~/.device-flow-cli/token.json` with restricted permissions (0600)
- CSRF protection using state parameter
- Automatic token expiry checking
- Local callback server only binds to localhost

## Development

### Build

```bash
go build -o device-flow-cli
```

### Run Tests

```bash
go test ./...
```

## License

MIT
