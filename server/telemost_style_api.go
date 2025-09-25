package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/mattermost/mattermost/server/public/plugin"
)

// ServeHTTP handles HTTP requests to the plugin
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	config := p.getConfiguration()
	if err := config.IsValid(); err != nil {
		http.Error(w, "This plugin is not configured.", http.StatusForbidden)
		return
	}

	path := r.URL.Path
	p.API.LogInfo("HTTP request received", "path", path)

	switch {
	case path == "/api/v1/meetings":
		p.handleCreateMeeting(w, r)
	case path == "/oauth/start":
		p.handleOAuthStart(w, r)
	case path == "/oauth/callback":
		p.handleOAuthCallback(w, r)
	case path == "/oauth/complete":
		p.handleOAuthComplete(w, r)
	case strings.HasPrefix(path, "/assets/"):
		p.handleAssets(w, r)
	case strings.Contains(path, "telemost-icon.svg"):
		// Handle direct icon requests
		p.handleAssets(w, r)
	default:
		p.API.LogInfo("No handler found for path", "path", path)
		http.NotFound(w, r)
	}
}

// handleAssets serves static assets from the plugin
func (p *Plugin) handleAssets(w http.ResponseWriter, r *http.Request) {
	// Debug logging
	p.API.LogInfo("handleAssets called", "path", r.URL.Path)

	var assetPath string

	// Handle different URL patterns
	if strings.HasPrefix(r.URL.Path, "/assets/") {
		// Pattern: /assets/telemost-icon.svg
		assetPath = strings.TrimPrefix(r.URL.Path, "/assets/")
	} else if strings.Contains(r.URL.Path, "telemost-icon.svg") {
		// Pattern: /plugins/com.mattermost.plugin-telemost/assets/telemost-icon.svg
		assetPath = "telemost-icon.svg"
	} else {
		p.API.LogError("Unknown asset path pattern", "path", r.URL.Path)
		http.NotFound(w, r)
		return
	}

	p.API.LogInfo("Extracted asset path", "assetPath", assetPath)

	// Security check - only allow files in assets directory
	if strings.Contains(assetPath, "..") || strings.Contains(assetPath, "/") {
		p.API.LogError("Security violation in asset path", "assetPath", assetPath)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Construct the full path to the asset
	fullPath := filepath.Join("assets", assetPath)
	p.API.LogInfo("Reading asset file", "fullPath", fullPath)

	// Read the asset file
	assetData, err := p.API.ReadFile(fullPath)
	if err != nil {
		p.API.LogError("Failed to read asset file", "fullPath", fullPath, "error", err.Error())
		http.NotFound(w, r)
		return
	}

	// Set appropriate content type based on file extension
	ext := strings.ToLower(filepath.Ext(assetPath))
	switch ext {
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	// Set cache headers
	w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year

	// Write the asset data
	w.Write(assetData)
}

// handleCreateMeeting handles meeting creation requests
func (p *Plugin) handleCreateMeeting(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Title       string   `json:"title"`
		Description string   `json:"description"`
		Cohosts     []string `json:"cohosts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user token from request context or headers
	// This is a simplified implementation - you may need to adjust based on your auth flow
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Create meeting using Telemost client
	meeting, err := p.telemostClient.CreateMeetingWithDefaults(p.getConfiguration(), req.Title, req.Description, req.Cohosts)
	if err != nil {
		p.API.LogError("Failed to create meeting", "error", err.Error())
		http.Error(w, "Failed to create meeting", http.StatusInternalServerError)
		return
	}

	// Return meeting details
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meeting)
}
