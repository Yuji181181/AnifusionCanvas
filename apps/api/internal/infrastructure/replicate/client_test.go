package replicate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreatePrediction(t *testing.T) {
	mock := newMockReplicateServer()
	defer mock.close()

	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   "test-token",
		baseURL:    mock.server.URL,
	}

	prediction, err := client.CreatePrediction(context.Background(), "test-model-version", map[string]any{
		"image": "data:image/png;base64,test",
	})
	if err != nil {
		t.Fatalf("create prediction failed: %v", err)
	}
	if prediction.ID == "" {
		t.Fatalf("expected prediction ID to be set")
	}
	if prediction.Status != StatusStarting {
		t.Fatalf("expected starting status, got %s", prediction.Status)
	}
}

func TestGetPrediction(t *testing.T) {
	mock := newMockReplicateServer()
	defer mock.close()

	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   "test-token",
		baseURL:    mock.server.URL,
	}

	prediction, err := client.CreatePrediction(context.Background(), "v1", map[string]any{"prompt": "test"})
	if err != nil {
		t.Fatalf("create prediction failed: %v", err)
	}

	mock.completePrediction(prediction.ID, "https://example.test/output.mp4")

	got, err := client.GetPrediction(context.Background(), prediction.ID)
	if err != nil {
		t.Fatalf("get prediction failed: %v", err)
	}
	if got.Status != StatusSucceeded {
		t.Fatalf("expected succeeded status, got %s", got.Status)
	}
	if got.Output != "https://example.test/output.mp4" {
		t.Fatalf("expected output URL, got %v", got.Output)
	}
}

func TestWaitForPrediction(t *testing.T) {
	mock := newMockReplicateServer()
	defer mock.close()

	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   "test-token",
		baseURL:    mock.server.URL,
	}

	prediction, err := client.CreatePrediction(context.Background(), "v1", map[string]any{"prompt": "test"})
	if err != nil {
		t.Fatalf("create prediction failed: %v", err)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		mock.completePrediction(prediction.ID, "https://example.test/result.png")
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	got, err := client.WaitForPrediction(ctx, prediction.ID, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("wait for prediction failed: %v", err)
	}
	if got.Status != StatusSucceeded {
		t.Fatalf("expected succeeded, got %s", got.Status)
	}
}

func TestWaitForPredictionTimeout(t *testing.T) {
	mock := newMockReplicateServer()
	defer mock.close()

	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   "test-token",
		baseURL:    mock.server.URL,
	}

	prediction, err := client.CreatePrediction(context.Background(), "v1", map[string]any{})
	if err != nil {
		t.Fatalf("create prediction failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = client.WaitForPrediction(ctx, prediction.ID, 10*time.Second)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
}

func TestDownloadOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("fake-video-data"))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   "test-token",
		baseURL:    "",
	}

	data, err := client.DownloadOutput(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("download failed: %v", err)
	}
	if string(data) != "fake-video-data" {
		t.Fatalf("unexpected download body: %q", string(data))
	}
}

func TestDownloadOutputError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("gone"))
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   "test-token",
	}

	_, err := client.DownloadOutput(context.Background(), server.URL)
	if err == nil {
		t.Fatalf("expected download error")
	}
}

func TestCreatePredictionNoToken(t *testing.T) {
	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   "",
		baseURL:    "http://localhost",
	}

	_, err := client.CreatePrediction(context.Background(), "v1", map[string]any{})
	if err == nil {
		t.Fatalf("expected error for missing token")
	}
}

func TestPredictionOutputURL(t *testing.T) {
	prediction := &Prediction{Output: "https://example.test/file.mp4"}
	if url := prediction.OutputURL(); url != "https://example.test/file.mp4" {
		t.Fatalf("unexpected URL: %q", url)
	}
}

func TestPredictionOutputURLList(t *testing.T) {
	prediction := &Prediction{Output: []any{"https://a.test/1.png", "https://a.test/2.png"}}
	urls := prediction.OutputURLList()
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
	if urls[0] != "https://a.test/1.png" || urls[1] != "https://a.test/2.png" {
		t.Fatalf("unexpected URLs: %v", urls)
	}
}

func TestAPIError(t *testing.T) {
	err := &APIError{StatusCode: 422, Body: `{"detail":"invalid input"}`}
	if err.Error() == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestPredictionStatusTerminal(t *testing.T) {
	if !StatusSucceeded.Terminal() || !StatusFailed.Terminal() || !StatusCanceled.Terminal() {
		t.Fatalf("expected terminal statuses to be terminal")
	}
	if StatusStarting.Terminal() || StatusProcessing.Terminal() {
		t.Fatalf("expected non-terminal statuses to not be terminal")
	}
}

func TestDownloadOutputServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, 1024)
		for i := range data {
			data[i] = 'x'
		}
		w.WriteHeader(http.StatusBadGateway)
		w.Write(data)
	}))
	defer server.Close()

	client := &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiToken:   "test-token",
	}
	_, err := client.DownloadOutput(context.Background(), server.URL)
	if err == nil {
		t.Fatalf("expected error for non-2xx download response")
	}
}
