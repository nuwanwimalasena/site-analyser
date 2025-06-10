package main

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/html"
)

// Mock HTTP Client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

// Mock HTML Parser
type MockHTMLParser struct {
	mock.Mock
}

func (m *MockHTMLParser) Parse(r io.Reader) (*html.Node, error) {
	args := m.Called(r)
	return args.Get(0).(*html.Node), args.Error(1)
}

// Test cases for ReadPageContent
func TestReadPageContent(t *testing.T) {
	// Ensure gock is activated and deactivated after the test
	defer gock.Off()

	tests := []struct {
		name          string
		url           string
		mockResponse  string
		mockStatus    int
		mockError     error
		expectedError string
		assertions    func(*testing.T, *PageAnalysis)
	}{
		{
			name:         "Valid URL with successful response",
			url:          "https://example.com",
			mockResponse: "<!DOCTYPE html><html><head><title>Test Page</title></head><body><h1>Hello</h1></body></html>",
			mockStatus:   200,
			assertions: func(t *testing.T, result *PageAnalysis) {
				assert.Equal(t, "HTML 5", result.HTMLVersion)
				assert.Equal(t, "Test Page", result.Title)
				assert.Len(t, result.Headings, 1)
				assert.Equal(t, "h1", result.Headings[0].Tag)
				assert.Equal(t, 1, result.Headings[0].Count)
			},
		},
		{
			name:          "Network error",
			url:           "https://example.com",
			mockError:     errors.New("connection refused"),
			expectedError: "connection refused",
		},
		{
			name:          "HTTP error status",
			url:           "https://example.com",
			mockStatus:    404,
			expectedError: "HTTP error: received status code 404",
		},
		{
			name:         "Page with login form",
			url:          "https://example.com/login",
			mockResponse: `<!DOCTYPE html><html><body><form><input type="text" name="username"><input type="password" name="password"><input type="submit" value="Login"></form></body></html>`,
			mockStatus:   200,
			assertions: func(t *testing.T, result *PageAnalysis) {
				assert.True(t, result.LoginForm)
			},
		},
		{
			name:         "Page without login form",
			url:          "https://example.com/nologin",
			mockResponse: `<!DOCTYPE html><html><body><form><input type="text" name="search"><input type="submit" value="Search"></form></body></html>`,
			mockStatus:   200,
			assertions: func(t *testing.T, result *PageAnalysis) {
				assert.False(t, result.LoginForm)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset gock for each test
			gock.Off()

			if tt.mockError != nil {
				gock.New(tt.url).ReplyError(tt.mockError)
			} else {
				gock.New(tt.url).Reply(tt.mockStatus).BodyString(tt.mockResponse)
			}

			result, err := ReadPageContent(tt.url)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tt.assertions != nil {
					tt.assertions(t, result)
				}
			}

			// Verify all mocks were called
			assert.True(t, gock.IsDone())
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		input   string
		isValid bool
	}{
		// Valid URLs
		{"http://example.com", true},
		{"https://example.com", true},
		{"https://sub.example.com/path", true},
		{"http://example.co.uk", true},
		{"https://example.com/abc/def", true},
		{"example.com", true},
		{"www.example.com", true},
		{"https://example.com/", true},
		// Invalid URLs
		{"htp://example.com", false},
		{"http:/example.com", false},
		{"example", false},
		{"http://", false},
		{"", false},
		{"http://.com", false},
		{"http://example", false},
		{"http://example.c", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ValidateURL(tt.input)
			assert.Equal(t, tt.isValid, result, "input: %s", tt.input)
		})
	}
}
