# Telemost Plugin for Mattermost

[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)](https://github.com/mattermost/mattermost-plugin-telemost)
[![Mattermost](https://img.shields.io/badge/Mattermost-6.2.1+-blue.svg)](https://mattermost.com)
[![Telemost](https://img.shields.io/badge/Telemost-API-green.svg)](https://telemost.yandex.com)

The Telemost plugin for Mattermost allows users to create and join Telemost video meetings directly from Mattermost channels. This plugin integrates with Yandex's Telemost video conferencing service, providing seamless video meeting capabilities within your Mattermost workspace.

## Features

- **🎥 Create Meetings**: Start Telemost video meetings with a simple slash command
- **🔐 OAuth Authentication**: Secure user authentication with Telemost using OAuth 2.0
- **💬 Channel Integration**: Meeting invitations appear as rich messages in channels
- **👥 Team Collaboration**: Share meeting links with your team instantly
- **⚡ Quick Access**: Fast meeting creation with `/telemost start` command

## Installation

### From Source

1. Clone this repository:
   ```bash
   git clone https://github.com/mattermost/mattermost-plugin-telemost.git
   cd mattermost-plugin-telemost
   ```

2. Build the plugin:
   ```bash
   make dist
   ```

3. Upload the plugin to your Mattermost server:
   - Go to System Console → Plugins → Plugin Management
   - Upload the `com.mattermost.plugin-telemost-*.tar.gz` file
   - Enable the plugin

### From Release

1. Download the latest release from the [releases page](https://github.com/mattermost/mattermost-plugin-telemost/releases)
2. Upload the plugin file to your Mattermost server
3. Enable the plugin in System Console

## Configuration

### System Console Settings

Navigate to **System Console → Plugins → Telemost** to configure:

#### Required Settings

- **Yandex OAuth Client ID**: Your Yandex application OAuth client ID
- **Mattermost Site URL**: Your Mattermost server URL (e.g., `https://mattermost.example.com`)

#### Optional Settings

- **Telemost OAuth Token (Legacy)**: Legacy server-side OAuth token for backward compatibility
- **Default Waiting Room Level**: 
  - `PUBLIC`: No waiting room (default)
  - `ORGANIZATION`: Waiting room for external users
  - `ADMINS`: Waiting room for all except organizers
- **Enable Live Stream**: Enable live streaming capability for meetings
- **Default Live Stream Access Level**:
  - `PUBLIC`: For all users
  - `ORGANIZATION`: Only for employees

### Yandex Cloud Setup

1. Go to [Yandex Cloud Console](https://cloud.yandex.com/)
2. Create a new OAuth application
3. Set the redirect URI to: `https://your-mattermost-server.com/plugins/com.mattermost.plugin-telemost/oauth/callback`
4. Copy the Client ID and configure it in the plugin settings

## Usage

### Slash Commands

The plugin provides several slash commands for managing Telemost meetings:

#### `/telemost start`
Creates a new Telemost meeting and posts an invitation to the channel.

**Requirements**: User must be authenticated with Telemost (see `/telemost connect`)

**Example**:
```
/telemost start
```

#### `/telemost connect`
Initiates OAuth authentication with Telemost.

**Example**:
```
/telemost connect
```

**Response**: Provides a link to complete OAuth authentication in your browser.

#### `/telemost disconnect`
Removes your Telemost authentication.

**Example**:
```
/telemost disconnect
```

#### `/telemost help`
Shows available commands and usage information.

**Example**:
```
/telemost help
```

### Meeting Invitations

When a meeting is created, the plugin posts a rich message to the channel containing:

- **Meeting Title**: "Telemost Meeting"
- **Join Button**: Direct link to join the meeting
- **Meeting ID**: Unique identifier for the meeting
- **Custom Icon**: Telemost branding

### User Authentication Flow

1. User runs `/telemost connect`
2. User clicks the authentication link
3. User completes OAuth flow in browser
4. User is redirected back to Mattermost
5. User can now create meetings with `/telemost start`

## Development

### Prerequisites

- Go 1.19+
- Node.js 16+
- npm 8+

### Building

```bash
# Install dependencies
make install-go-tools
cd webapp && npm install

# Build the plugin
make dist

# Run tests
make test

# Check code style
make check-style
```

### Development Mode

```bash
# Watch for changes and rebuild
make watch

# Deploy to local server
make deploy
```

### Project Structure

```
├── server/                 # Go server-side code
│   ├── command/            # Slash command handlers
│   ├── store/             # Data storage utilities
│   └── *.go               # Core plugin logic
├── webapp/                 # React frontend code
│   ├── src/               # React components
│   └── dist/              # Built webapp assets
├── assets/                # Plugin assets (icons, etc.)
├── public/                # Static web assets
└── build/                 # Build utilities
```

## API Reference

### Server Endpoints

- `POST /api/v1/meetings` - Create a new meeting
- `GET /oauth/start` - Start OAuth authentication
- `GET /oauth/callback` - OAuth callback handler
- `GET /oauth/complete` - Complete OAuth flow

### Data Models

#### TelemostMeeting
```go
type TelemostMeeting struct {
    ID         string `json:"id"`
    JoinURL    string `json:"join_url"`
    LiveStream *struct {
        WatchURL string `json:"watch_url"`
    } `json:"live_stream,omitempty"`
}
```

#### UserToken
```go
type UserToken struct {
    AccessToken string `json:"access_token"`
    ExpiresAt   string `json:"expires_at"`
    UserID      string `json:"user_id"`
}
```

## Troubleshooting

### Common Issues

#### "Telemost not authenticated!" Error
- **Cause**: User hasn't completed OAuth authentication
- **Solution**: Run `/telemost connect` and complete the OAuth flow

#### "Failed to create meeting!" Error
- **Cause**: Invalid OAuth token or API error
- **Solution**: Try `/telemost disconnect` then `/telemost connect` to re-authenticate

#### OAuth Redirect Issues
- **Cause**: Incorrect Site URL configuration
- **Solution**: Verify the Site URL in System Console matches your server URL

### Debug Mode

Enable debug logging by setting `MM_DEBUG=1` when building:

```bash
MM_DEBUG=1 make dist
```

### Logs

View plugin logs:
```bash
make logs
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Run tests: `make test`
5. Commit your changes: `git commit -m "Add feature"`
6. Push to the branch: `git push origin feature-name`
7. Submit a pull request

## License

This plugin is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [Plugin Documentation](https://developers.mattermost.com/extend/plugins/)
- **Issues**: [GitHub Issues](https://github.com/mattermost/mattermost-plugin-telemost/issues)
- **Community**: [Mattermost Community](https://community.mattermost.com)

## Changelog

### v0.5.0
- Initial release
- OAuth authentication support
- Meeting creation and management
- Rich meeting invitations
- Custom Telemost branding

---

**Made with ❤️ for the Mattermost community**