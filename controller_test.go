package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Load templates from the correct path
	r.LoadHTMLGlob("templates/*")

	r.GET("/", FormHandler)
	r.POST("/results", ResultsHandler)
	return r
}

func TestFormHandler(t *testing.T) {
	router := setupTestRouter()

	// Test GET request to form handler
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, req)

	// Print the full response for debugging
	fmt.Printf("\nForm Handler Response:\n%s\n", w.Body.String())

	// Check status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check content type
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")

	// Check for specific form elements that should be in the template
	body := w.Body.String()
	assert.Contains(t, body, "<title>Site Analyser</title>")
	assert.Contains(t, body, "action=\"/results\"")
	assert.Contains(t, body, "method=\"post\"")
	assert.Contains(t, body, "name=\"url\"")
	assert.Contains(t, body, "class=\"form-control\"")
	assert.Contains(t, body, "value=\"Analyse\"")
}

func TestResultsHandler(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid URL",
			url:            "invalid-url",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid URL",
		},
		{
			name:           "Valid URL",
			url:            "https://example.com",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/results", strings.NewReader("url="+tt.url))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			router.ServeHTTP(w, req)

			// Print the full response for debugging
			fmt.Printf("\nResults Handler Response for %s:\n%s\n", tt.name, w.Body.String())

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "text/html")

			body := w.Body.String()
			if tt.expectedError != "" {
				assert.Contains(t, body, tt.expectedError)
				// For error case, verify we're back at the form
				assert.Contains(t, body, "<title>Site Analyser</title>")
				assert.Contains(t, body, "class=\"alert alert-danger small w-100\"")
			} else {
				// For success case, verify we're on the results page
				assert.Contains(t, body, tt.url)
				assert.Contains(t, body, "Site Analyser Results")
			}
		})
	}
}
