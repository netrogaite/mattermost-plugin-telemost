package command

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

const telemostCommandTrigger = "telemost"

// UserToken represents a user's OAuth token
type UserToken struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   string `json:"expires_at"`
	UserID      string `json:"user_id"`
}

// TelemostMeeting represents a Telemost meeting
type TelemostMeeting struct {
	ID         string `json:"id"`
	JoinURL    string `json:"join_url"`
	LiveStream *struct {
		WatchURL string `json:"watch_url"`
	} `json:"live_stream,omitempty"`
}

// RegisterCommand registers the telemost slash command
func RegisterCommand(client *pluginapi.Client) error {
	// Base64 encoded Telemost icon SVG with data URL prefix
	iconData := "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMzIiIGhlaWdodD0iMzIiIGZpbGw9Im5vbmUiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHBhdGggZD0iTTAgMTZDMCA5LjQ0IDAgNi4xNiAxLjYzIDMuODRhOSA5IDAgMCAxIDIuMi0yLjIxQzYuMTcgMCA5LjQ1IDAgMTYgMHM5Ljg0IDAgMTIuMTYgMS42M2E5IDkgMCAwIDEgMi4yMSAyLjJDMzIgNi4xNyAzMiA5LjQ1IDMyIDE2czAgOS44NC0xLjYzIDEyLjE2YTkgOSAwIDAgMS0yLjIgMi4yMUMyNS44MyAzMiAyMi41NSAzMiAxNiAzMnMtOS44NCAwLTEyLjE2LTEuNjNhOSA5IDAgMCAxLTIuMjEtMi4yQzAgMjUuODMgMCAyMi41NSAwIDE2eiIgZmlsbD0idXJsKCNhKSIvPjxjaXJjbGUgb3BhY2l0eT0iLjYiIGN4PSIxNC4xNyIgY3k9IjEzLjUiIHI9IjEuNSIgZmlsbD0iI2ZmZiIvPjxjaXJjbGUgY3g9IjEyIiBjeT0iMTYiIHI9IjciIHN0cm9rZT0iI2ZmZiIgc3Ryb2tlLXdpZHRoPSIxLjUiLz48Y2lyY2xlIGN4PSIyNSIgY3k9IjE2IiBmaWxsPSIjMEYwIiByPSIzIi8+PGRlZnM+PGxpbmVhckdyYWRpZW50IGlkPSJhIiB4MT0iLTMuMjciIHkxPSIzNS42NCIgeDI9IjMyIiB5Mj0iMCIgZ3JhZGllbnRVbml0cz0idXNlclNwYWNlT25Vc2UiPjxzdG9wLz48c3RvcCBvZmZzZXQ9IjEiIHN0b3AtY29sb3I9IiMzMTQwNEUiLz48L2xpbmVhckdyYWRpZW50PjwvZGVmcz48L3N2Zz4="

	cmd := &model.Command{
		Trigger:              telemostCommandTrigger,
		AutoComplete:         true,
		AutoCompleteDesc:     "Available commands: start | connect | disconnect | help",
		AutoCompleteHint:     "[command]",
		AutocompleteIconData: iconData,
		AutocompleteData:     model.NewAutocompleteData(telemostCommandTrigger, "[command]", "Available commands: start | connect | disconnect | help"),
	}

	client.Log.Info("RegisterCommand: About to register command", "trigger", telemostCommandTrigger)

	err := client.SlashCommand.Register(cmd)
	if err != nil {
		client.Log.Error("RegisterCommand: Failed to register command", "error", err)
		return err
	}

	client.Log.Info("RegisterCommand: Command registered successfully")
	return nil
}
