package mcputils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// SQClient is a SonarQube API HTTP client.
type SQClient struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewSQClient creates a new SonarQube API client using configured URL and token.
func NewSQClient() *SQClient {
	configOnce.Do(initConfig)
	return &SQClient{
		BaseURL: strings.TrimSuffix(sonarQubeURL, "/"),
		Token:   sqToken,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

// NewSQClientWithToken creates a client with an explicit token (for HTTP per-request auth).
func NewSQClientWithToken(baseURL, token string) *SQClient {
	return &SQClient{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		Token:   token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
		},
	}
}

// DoGet performs a GET request to the SonarQube API and decodes the JSON response.
func (c *SQClient) DoGet(ctx context.Context, path string, params url.Values, result interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, params, nil, result)
}

// DoPost performs a POST request to the SonarQube API and decodes the JSON response.
func (c *SQClient) DoPost(ctx context.Context, path string, params url.Values, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, params, nil, result)
}

// DoPostWithBody performs a POST request with a JSON body.
func (c *SQClient) DoPostWithBody(ctx context.Context, path string, params url.Values, body interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, params, body, result)
}

// DoGetRaw performs a GET and returns the raw response body as string.
func (c *SQClient) DoGetRaw(ctx context.Context, path string, params url.Values) (string, error) {
	u := c.BaseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	if c.Token != "" {
		req.SetBasicAuth(c.Token, "")
	}
	req.Header.Set("Accept", "text/plain, application/json")

	LogRequest(http.MethodGet, u, map[string][]string(params), req.Header, nil)
	start := time.Now()
	resp, err := c.Client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	LogResponse(ctx, resp.StatusCode, http.MethodGet, u, time.Since(start), body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return string(body), nil
}

func (c *SQClient) doRequest(ctx context.Context, method, path string, params url.Values, body interface{}, result interface{}) error {
	u := c.BaseURL + path

	var bodyReader io.Reader
	var bodyBytes []byte

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyBytes = data
		bodyReader = strings.NewReader(string(data))
	}

	if method == http.MethodGet && len(params) > 0 {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if c.Token != "" {
		req.SetBasicAuth(c.Token, "")
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	if method == http.MethodPost && len(params) > 0 {
		req.URL.RawQuery = params.Encode()
	}

	LogRequest(method, u, map[string][]string(params), req.Header, bodyBytes)
	start := time.Now()
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	LogResponse(ctx, resp.StatusCode, method, u, time.Since(start), respBody)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

// ResolveProjectKey returns the project key from args if provided, otherwise the configured default.
// Returns an error if neither is available.
func ResolveProjectKey(args map[string]interface{}, paramName string) (string, error) {
	if v, ok := args[paramName]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s, nil
		}
	}
	defaultKey := GetDefaultProjectKey()
	if defaultKey != "" {
		return defaultKey, nil
	}
	return "", fmt.Errorf("%s is required (no default project key configured)", paramName)
}

// OptionalProjectKey returns the project key if available, or empty string.
func OptionalProjectKey(args map[string]interface{}, paramName string) string {
	if v, ok := args[paramName]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return GetDefaultProjectKey()
}

// GetOptionalString extracts an optional string argument.
func GetOptionalString(args map[string]interface{}, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// GetStringArray extracts an optional string array argument.
// Accepts MCP JSON arrays ([]interface{}), []string, or a single string
// (CLI mode passes --files as string; comma-separated values are supported).
func GetStringArray(args map[string]interface{}, key string) []string {
	v, ok := args[key]
	if !ok {
		return nil
	}

	switch val := v.(type) {
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, s)
			}
		}
		return result
	case []string:
		result := make([]string, 0, len(val))
		for _, s := range val {
			if s != "" {
				result = append(result, s)
			}
		}
		return result
	case string:
		if val == "" {
			return nil
		}
		parts := strings.Split(val, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				result = append(result, s)
			}
		}
		return result
	default:
		return nil
	}
}

// GetIntOrDefault extracts an integer argument with a default value.
func GetIntOrDefault(args map[string]interface{}, key string, defaultVal int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return defaultVal
}

// GetBoolOrDefault extracts a boolean argument with a default value.
func GetBoolOrDefault(args map[string]interface{}, key string, defaultVal bool) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return defaultVal
}

// BuildProjectKeyParams adds projectKey, branch, and pullRequest to query params.
func BuildProjectKeyParams(args map[string]interface{}, params url.Values) url.Values {
	if projectKey := OptionalProjectKey(args, "projectKey"); projectKey != "" {
		params.Set("project", projectKey)
	}
	if branch := GetOptionalString(args, "branch"); branch != "" {
		params.Set("branch", branch)
	}
	if pr := GetOptionalString(args, "pullRequest"); pr != "" {
		params.Set("pullRequest", pr)
	}
	return params
}
