package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mattermost/mattermost/server/public/plugin"
)

const (
	telemostAPIBaseURL = "https://cloud-api.yandex.net/v1/telemost-api"
)

// TelemostMeeting represents a Telemost meeting response
type TelemostMeeting struct {
	ID         string `json:"id"`
	JoinURL    string `json:"join_url"`
	LiveStream *struct {
		WatchURL string `json:"watch_url"`
	} `json:"live_stream,omitempty"`
}

// TelemostCreateRequest represents the request body for creating a meeting
type TelemostCreateRequest struct {
	WaitingRoomLevel string `json:"waiting_room_level,omitempty"`
	LiveStream       *struct {
		AccessLevel string `json:"access_level,omitempty"`
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"live_stream,omitempty"`
	Cohosts []struct {
		Email string `json:"email"`
	} `json:"cohosts,omitempty"`
}

// TelemostError represents an error response from the Telemost API
type TelemostError struct {
	Error       string `json:"error"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Details     struct {
		Emails string `json:"emails,omitempty"`
	} `json:"details,omitempty"`
}

// TelemostClient handles API interactions with Telemost
type TelemostClient struct {
	oauthToken string
	api        plugin.API
}

// NewTelemostClient creates a new Telemost API client
func NewTelemostClient(oauthToken string, api plugin.API) *TelemostClient {
	return &TelemostClient{
		oauthToken: oauthToken,
		api:        api,
	}
}

// CreateMeeting creates a new Telemost meeting
func (tc *TelemostClient) CreateMeeting(req *TelemostCreateRequest) (*TelemostMeeting, error) {
	url := fmt.Sprintf("%s/conferences", telemostAPIBaseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "OAuth "+tc.oauthToken)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		var telemostErr TelemostError
		if err := json.Unmarshal(body, &telemostErr); err != nil {
			return nil, fmt.Errorf("failed to create meeting, status: %d, body: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("telemost API error: %s - %s", telemostErr.Error, telemostErr.Message)
	}

	var meeting TelemostMeeting
	if err := json.Unmarshal(body, &meeting); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &meeting, nil
}

// CreateMeetingWithDefaults creates a meeting with default settings from configuration
func (tc *TelemostClient) CreateMeetingWithDefaults(config interface {
	GetDefaultWaitingRoomLevel() string
	IsLiveStreamEnabled() bool
	GetDefaultLiveStreamAccessLevel() string
}, title string, description string, cohosts []string) (*TelemostMeeting, error) {
	req := &TelemostCreateRequest{
		WaitingRoomLevel: config.GetDefaultWaitingRoomLevel(),
	}

	// Add live stream if enabled
	if config.IsLiveStreamEnabled() {
		req.LiveStream = &struct {
			AccessLevel string `json:"access_level,omitempty"`
			Title       string `json:"title,omitempty"`
			Description string `json:"description,omitempty"`
		}{
			AccessLevel: config.GetDefaultLiveStreamAccessLevel(),
			Title:       title,
			Description: description,
		}
	}

	// Add cohosts if provided
	if len(cohosts) > 0 {
		req.Cohosts = make([]struct {
			Email string `json:"email"`
		}, len(cohosts))
		for i, email := range cohosts {
			req.Cohosts[i].Email = email
		}
	}

	return tc.CreateMeeting(req)
}
