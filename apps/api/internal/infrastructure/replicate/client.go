package replicate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.replicate.com/v1"

// Client provides access to the Replicate API for creating and monitoring predictions.
type Client struct {
	httpClient *http.Client
	apiToken   string
	baseURL    string
}

// NewClient creates a new Replicate API client.
func NewClient(apiToken string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiToken:   apiToken,
		baseURL:    defaultBaseURL,
	}
}

// CreatePrediction starts a new prediction on Replicate.
func (c *Client) CreatePrediction(ctx context.Context, version string, input map[string]any) (*Prediction, error) {
	if c.apiToken == "" {
		return nil, fmt.Errorf("replicate API token is not set")
	}

	body := CreatePredictionRequest{
		Version: version,
		Input:   input,
	}

	var prediction Prediction
	if err := c.do(ctx, http.MethodPost, "/predictions", body, &prediction); err != nil {
		return nil, fmt.Errorf("create prediction: %w", err)
	}

	return &prediction, nil
}

// GetPrediction retrieves the current state of a prediction.
func (c *Client) GetPrediction(ctx context.Context, id string) (*Prediction, error) {
	var prediction Prediction
	if err := c.do(ctx, http.MethodGet, "/predictions/"+id, nil, &prediction); err != nil {
		return nil, fmt.Errorf("get prediction: %w", err)
	}
	return &prediction, nil
}

// WaitForPrediction polls the prediction until it reaches a terminal state.
// pollInterval controls the time between polls. The context deadline controls
// the maximum wait time.
func (c *Client) WaitForPrediction(ctx context.Context, id string, pollInterval time.Duration) (*Prediction, error) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		prediction, err := c.GetPrediction(ctx, id)
		if err != nil {
			return nil, err
		}

		if prediction.Status.Terminal() {
			return prediction, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
		}
	}
}

// DownloadOutput downloads the file at the given URL. Used to fetch prediction
// output files (images, videos) from Replicate's temporary URLs.
func (c *Client) DownloadOutput(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build download request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download output: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("download returned status %d: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read download body: %w", err)
	}

	return data, nil
}

func (c *Client) do(ctx context.Context, method string, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{
			StatusCode: resp.StatusCode,
			Body:       string(respBody),
		}
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return nil
}

// APIError represents an error response from the Replicate API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("replicate API error status=%d body=%s", e.StatusCode, e.Body)
}
