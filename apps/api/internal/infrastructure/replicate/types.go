package replicate

import "time"

// PredictionStatus represents the status of a Replicate prediction.
type PredictionStatus string

const (
	StatusStarting   PredictionStatus = "starting"
	StatusProcessing PredictionStatus = "processing"
	StatusSucceeded  PredictionStatus = "succeeded"
	StatusFailed     PredictionStatus = "failed"
	StatusCanceled   PredictionStatus = "canceled"
)

// Terminal returns true if the prediction has reached a terminal state.
func (s PredictionStatus) Terminal() bool {
	return s == StatusSucceeded || s == StatusFailed || s == StatusCanceled
}

// CreatePredictionRequest is the request body for POST /v1/predictions.
type CreatePredictionRequest struct {
	Version string         `json:"version"`
	Input   map[string]any `json:"input"`
}

// Prediction is the response from the Replicate predictions API.
type Prediction struct {
	ID        string           `json:"id"`
	Version   string           `json:"version"`
	Status    PredictionStatus `json:"status"`
	Input     map[string]any   `json:"input"`
	Output    any              `json:"output"`
	Error     string           `json:"error,omitempty"`
	Logs      string           `json:"logs,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	StartedAt *time.Time       `json:"started_at,omitempty"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
}

// OutputURLs extracts string URLs from the prediction output.
// Replicate typically returns a single URL string or []string for models
// that produce multiple output files.
func (p *Prediction) OutputURL() string {
	if s, ok := p.Output.(string); ok {
		return s
	}
	return ""
}

// OutputURLList extracts a list of string URLs from the prediction output.
func (p *Prediction) OutputURLList() []string {
	switch v := p.Output.(type) {
	case []any:
		urls := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				urls = append(urls, s)
			}
		}
		return urls
	case []string:
		return v
	}
	return nil
}
