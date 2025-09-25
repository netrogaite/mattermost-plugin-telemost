package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/mattermost/mattermost-plugin-starter-template/server/command"
	"github.com/mattermost/mattermost-plugin-starter-template/server/store/kvstore"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/pluginapi"
	"github.com/mattermost/mattermost/server/public/pluginapi/cluster"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// kvstore is the client used to read/write KV records for this plugin.
	kvstore kvstore.KVStore

	// client is the Mattermost server API client.
	client *pluginapi.Client

	// commandClient is the client used to register and execute slash commands.
	commandClient *command.Handler

	// telemostClient is the client used to interact with Telemost API.
	telemostClient *TelemostClient

	backgroundJob *cluster.Job

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration
}

// OnActivate is invoked when the plugin is activated. If an error is returned, the plugin will be deactivated.
func (p *Plugin) OnActivate() error {
	p.client = pluginapi.NewClient(p.API, p.Driver)

	// p.kvstore = kvstore.NewKVStore(p.client) // Commented out as NewKVStore doesn't exist

	// Register the telemost command
	if err := command.RegisterCommand(p.client); err != nil {
		return errors.Wrap(err, "failed to register command")
	}

	p.commandClient = command.NewCommandHandler(p.client, p)

	// Initialize Telemost client if configuration is available
	config := p.getConfiguration()
	if config != nil && config.TelemostOAuthToken != "" {
		p.telemostClient = NewTelemostClient(config.TelemostOAuthToken, p.API)
	}

	job, err := cluster.Schedule(
		p.API,
		"BackgroundJob",
		cluster.MakeWaitForRoundedInterval(1*time.Hour),
		p.runJob,
	)
	if err != nil {
		return errors.Wrap(err, "failed to schedule background job")
	}

	p.backgroundJob = job

	return nil
}

// OnDeactivate is invoked when the plugin is deactivated.
func (p *Plugin) OnDeactivate() error {
	if p.backgroundJob != nil {
		if err := p.backgroundJob.Close(); err != nil {
			p.API.LogError("Failed to close background job", "err", err)
		}
	}
	return nil
}

// This will execute the commands that were registered in the RegisterCommand function.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	response, appErr := p.commandClient.Handle(args)
	if appErr != nil {
		return nil, appErr
	}
	return response, nil
}

// See https://developers.mattermost.com/extend/plugins/server/reference/

// GetUserTokenForCommand retrieves a user's OAuth token for command handler
func (p *Plugin) GetUserTokenForCommand(userID string) (*command.UserToken, error) {
	tokenJSON, appErr := p.API.KVGet(userTokenKeyPrefix + userID)
	if appErr != nil {
		return nil, appErr
	}

	var userToken UserToken
	if err := json.Unmarshal(tokenJSON, &userToken); err != nil {
		return nil, err
	}

	// Check if token has expired
	if time.Now().After(userToken.ExpiresAt) {
		// Delete expired token
		p.API.KVDelete(userTokenKeyPrefix + userID)
		return nil, fmt.Errorf("token expired")
	}

	// Convert to command.UserToken
	return &command.UserToken{
		AccessToken: userToken.AccessToken,
		ExpiresAt:   userToken.ExpiresAt.Format(time.RFC3339),
		UserID:      userToken.UserID,
	}, nil
}

// CreateMeetingWithUserToken creates a meeting using user's OAuth token for command handler
func (p *Plugin) CreateMeetingWithUserToken(token, title, description string) (*command.TelemostMeeting, error) {
	// Create Telemost client with user's OAuth token
	client := NewTelemostClient(token, p.API)

	// Create the meeting
	meeting, err := client.CreateMeetingWithDefaults(p.getConfiguration(), title, description, []string{})
	if err != nil {
		return nil, err
	}

	// Convert to command.TelemostMeeting
	result := &command.TelemostMeeting{
		ID:      meeting.ID,
		JoinURL: meeting.JoinURL,
	}
	if meeting.LiveStream != nil {
		result.LiveStream = &struct {
			WatchURL string `json:"watch_url"`
		}{
			WatchURL: meeting.LiveStream.WatchURL,
		}
	}

	return result, nil
}

// runJob is a background job that runs periodically
func (p *Plugin) runJob() {
	// Include job logic here
	p.API.LogInfo("Background job is currently running")
}
