package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	oauthStateKeyPrefix = "telemost_oauth_state_"
	userTokenKeyPrefix  = "telemost_user_token_"
	oauthRedirectURL    = "/plugins/com.mattermost.plugin-telemost/oauth/callback"
)

// OAuthState represents the OAuth state for security
type OAuthState struct {
	UserID    string    `json:"user_id"`
	ChannelID string    `json:"channel_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// UserToken represents a user's OAuth token
type UserToken struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	UserID      string    `json:"user_id"`
}

// OAuthError represents an OAuth error response
type OAuthError struct {
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

// handleOAuthStart initiates the OAuth flow
func (p *Plugin) handleOAuthStart(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("Mattermost-User-Id")
	if userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get channel ID from query parameter
	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		http.Error(w, "Missing channel_id parameter", http.StatusBadRequest)
		return
	}

	// Generate OAuth state
	stateBytes := make([]byte, 32)
	if _, err := rand.Read(stateBytes); err != nil {
		p.API.LogError("Failed to generate OAuth state", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	state := base64.URLEncoding.EncodeToString(stateBytes)

	// Store OAuth state
	oauthState := OAuthState{
		UserID:    userID,
		ChannelID: channelID,
		ExpiresAt: time.Now().Add(10 * time.Minute), // 10 minute expiry
	}

	stateJSON, err := json.Marshal(oauthState)
	if err != nil {
		p.API.LogError("Failed to marshal OAuth state", "error", err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Store state in KV store
	appErr := p.API.KVSet(oauthStateKeyPrefix+state, stateJSON)
	if appErr != nil {
		p.API.LogError("Failed to store OAuth state", "error", appErr.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Build OAuth URL
	config := p.getConfiguration()
	redirectURI := config.SiteURL + oauthRedirectURL
	oauthURL := fmt.Sprintf(
		"https://oauth.yandex.ru/authorize?response_type=token&client_id=%s&redirect_uri=%s&state=%s",
		config.YandexClientID,
		url.QueryEscape(redirectURI),
		state,
	)

	// Redirect to Yandex OAuth
	http.Redirect(w, r, oauthURL, http.StatusTemporaryRedirect)
}

// handleOAuthCallback handles the OAuth callback from Yandex
func (p *Plugin) handleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	// For Yandex OAuth, we need to handle this differently since it uses URL fragments
	// The actual OAuth response will be in the fragment, but we can't access it server-side
	// Instead, we'll create a simple HTML page that extracts the token from the fragment

	// Get configuration to use proper Site URL
	config := p.getConfiguration()
	siteURL := config.SiteURL

	// Create HTML page that will extract the token from the URL fragment
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Telemost OAuth</title>
</head>
<body>
    <script>
        // Extract token from URL fragment
        const fragment = window.location.hash.substring(1);
        const params = new URLSearchParams(fragment);
        
        const accessToken = params.get('access_token');
        const error = params.get('error');
        const state = params.get('state');
        
        if (error) {
            const errorDesc = params.get('error_description');
            document.body.innerHTML = '<h1>OAuth Error</h1><p>Error: ' + error + '</p><p>Description: ' + errorDesc + '</p>';
        } else if (accessToken && state) {
            // Send token to our callback endpoint
            fetch('/plugins/com.mattermost.plugin-telemost/oauth/complete', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    access_token: accessToken,
                    state: state
                })
            }).then(response => {
                if (response.ok) {
                    document.body.innerHTML = '<h1>Success!</h1><p>Telemost has been configured successfully. Redirecting back to Mattermost...</p>';
                    // Redirect back to Mattermost main page
                    setTimeout(() => {
                        // Redirect to main Site URL
                        window.location.href = '%s';
                    }, 2000);
                } else {
                    document.body.innerHTML = '<h1>Error</h1><p>Failed to complete OAuth setup.</p>';
                }
            }).catch(err => {
                document.body.innerHTML = '<h1>Error</h1><p>Failed to complete OAuth setup: ' + err.message + '</p>';
            });
        } else {
            document.body.innerHTML = '<h1>Error</h1><p>Missing access token or state. Fragment: ' + fragment + '</p>';
        }
    </script>
</body>
</html>
`, siteURL)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// OAuthCompleteRequest represents the request to complete OAuth
type OAuthCompleteRequest struct {
	AccessToken string `json:"access_token"`
	State       string `json:"state"`
}

// handleOAuthComplete handles the completion of OAuth flow
func (p *Plugin) handleOAuthComplete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var req OAuthCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Retrieve and validate OAuth state
	stateJSON, appErr := p.API.KVGet(oauthStateKeyPrefix + req.State)
	if appErr != nil {
		p.API.LogError("Failed to retrieve OAuth state", "error", appErr.Error())
		http.Error(w, "Invalid or expired OAuth state", http.StatusBadRequest)
		return
	}

	var oauthState OAuthState
	if err := json.Unmarshal(stateJSON, &oauthState); err != nil {
		p.API.LogError("Failed to unmarshal OAuth state", "error", err.Error())
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	// Check if state has expired
	if time.Now().After(oauthState.ExpiresAt) {
		p.API.LogError("OAuth state expired", "user_id", oauthState.UserID)
		http.Error(w, "OAuth state expired", http.StatusBadRequest)
		return
	}

	// Store user token
	userToken := UserToken{
		AccessToken: req.AccessToken,
		ExpiresAt:   time.Now().Add(24 * time.Hour), // Assume 24 hour expiry
		UserID:      oauthState.UserID,
	}

	tokenJSON, marshalErr := json.Marshal(userToken)
	if marshalErr != nil {
		p.API.LogError("Failed to marshal user token", "error", marshalErr.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	storeAppErr := p.API.KVSet(userTokenKeyPrefix+oauthState.UserID, tokenJSON)
	if storeAppErr != nil {
		p.API.LogError("Failed to store user token", "error", storeAppErr.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Clean up OAuth state
	p.API.KVDelete(oauthStateKeyPrefix + req.State)

	// Post success message to the channel
	successPost := &model.Post{
		ChannelId: oauthState.ChannelID,
		Message:   "✅ **Connected to Telemost**\n\nYour Telemost authentication has been connected. Use `/telemost start` to create a meeting.",
		UserId:    oauthState.UserID,
	}

	if _, appErr := p.API.CreatePost(successPost); appErr != nil {
		p.API.LogError("Failed to create success post", "error", appErr.Error())
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "OAuth token stored successfully",
	})
}

// getUserToken retrieves a user's OAuth token
func (p *Plugin) getUserToken(userID string) (*UserToken, error) {
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

	return &userToken, nil
}

// isUserAuthenticated checks if a user has a valid OAuth token
func (p *Plugin) isUserAuthenticated(userID string) bool {
	_, err := p.getUserToken(userID)
	return err == nil
}
