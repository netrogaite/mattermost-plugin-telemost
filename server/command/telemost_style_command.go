package command

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/pluginapi"
)

// Handler handles slash commands
type Handler struct {
	client *pluginapi.Client
	plugin interface {
		CreateMeetingWithUserToken(token, title, description string) (*TelemostMeeting, error)
	}
}

// NewCommandHandler creates a new command handler
func NewCommandHandler(client *pluginapi.Client, plugin interface {
	CreateMeetingWithUserToken(token, title, description string) (*TelemostMeeting, error)
}) *Handler {
	return &Handler{
		client: client,
		plugin: plugin,
	}
}

// Handle handles slash command execution
func (h *Handler) Handle(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	// Parse command
	fields := strings.Fields(args.Command)

	// If no subcommand provided, show help
	if len(fields) < 2 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Available commands:**\n- `/telemost start` - Start a new meeting (requires authentication)\n- `/telemost connect` - Authenticate with Telemost OAuth\n- `/telemost disconnect` - Remove Telemost authentication\n- `/telemost help` - Show this help message",
		}, nil
	}

	subcommand := strings.ToLower(fields[1])

	switch subcommand {
	case "start":
		// Check if user has OAuth token (production logic)
		var userToken []byte
		err := h.client.KV.Get(fmt.Sprintf("telemost_user_token_%s", args.UserId), &userToken)
		if err != nil || userToken == nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "**Telemost not authenticated!**\n\nPlease authenticate with Telemost first:\n1. Use `/telemost connect` to start OAuth authentication\n2. Complete the OAuth flow in your browser\n3. Try `/telemost start` again",
			}, nil
		}

		// Parse user token
		var tokenData map[string]interface{}
		if err := json.Unmarshal(userToken, &tokenData); err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "**Authentication error!** Please reconnect with `/telemost connect`.",
			}, nil
		}

		accessToken, ok := tokenData["access_token"].(string)
		if !ok || accessToken == "" {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "**Invalid token!** Please reconnect with `/telemost connect`.",
			}, nil
		}

		// Create real Telemost meeting using production API
		meeting, err := h.plugin.CreateMeetingWithUserToken(accessToken, "Telemost Meeting", "")
		if err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         fmt.Sprintf("**❌ Failed to create meeting!**\n\nError: %s\n\nPlease try again or contact support.", err.Error()),
			}, nil
		}

		// Create custom post with props for the custom component to render
		response := &model.CommandResponse{
			ResponseType: model.CommandResponseTypeInChannel,
			Text:         "", // Empty text since we render everything custom
			Props: map[string]interface{}{
				"type":      "custom_telemost_meeting",
				"joinURL":   meeting.JoinURL,
				"meetingID": meeting.ID,
				"title":     "Telemost Meeting",
			},
		}
		return response, nil

	case "connect":
		// Check if user is already authenticated
		var userToken []byte
		err := h.client.KV.Get(fmt.Sprintf("telemost_user_token_%s", args.UserId), &userToken)
		if err == nil && userToken != nil {
			// User is already authenticated
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "**✅ Already Connected to Telemost**\n\nYou are already authenticated with Telemost. You can:\n- Use `/telemost start` to create a meeting\n- Use `/telemost disconnect` to remove authentication",
			}, nil
		}

		// Start OAuth authentication flow (production logic)
		// Get the site URL from configuration to build the full OAuth URL
		// For now, use a relative path that will be resolved by the browser
		oauthURL := fmt.Sprintf("/plugins/com.mattermost.plugin-telemost/oauth/start?channel_id=%s", args.ChannelId)

		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("**🔗 Telemost Authentication Required**\n\nTo connect to Telemost, please complete the OAuth authentication:\n\n[**Click here to authenticate with Telemost**](%s)\n\nAfter authentication, you'll be able to create meetings using `/telemost start`.", oauthURL),
		}, nil

	case "disconnect":
		// Check if user is authenticated before trying to disconnect
		var userToken []byte
		err := h.client.KV.Get(fmt.Sprintf("telemost_user_token_%s", args.UserId), &userToken)
		if err != nil || userToken == nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "**Not authenticated!** You are not currently authenticated with Telemost. Use `/telemost connect` to authenticate first.",
			}, nil
		}

		// Remove OAuth token (production logic)
		tokenKey := fmt.Sprintf("telemost_user_token_%s", args.UserId)
		err = h.client.KV.Delete(tokenKey)
		if err != nil {
			return &model.CommandResponse{
				ResponseType: model.CommandResponseTypeEphemeral,
				Text:         "**❌ Failed to disconnect!** There was an error removing your authentication. Please try again.",
			}, nil
		}

		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**✅ Disconnected from Telemost**\n\nYour Telemost authentication has been removed. Use `/telemost connect` to authenticate again.",
		}, nil

	case "help":
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "**Available commands:**\n- `/telemost start` - Start a new meeting (requires authentication)\n- `/telemost connect` - Authenticate with Telemost OAuth\n- `/telemost disconnect` - Remove Telemost authentication\n- `/telemost help` - Show this help message",
		}, nil

	default:
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         fmt.Sprintf("Unknown command: `%s`. Use `/telemost help` to see available commands.", subcommand),
		}, nil
	}
}
