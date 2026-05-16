package replicate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// mockReplicateServer provides a test HTTP server that mimics the Replicate API.
type mockReplicateServer struct {
	server      *httptest.Server
	mu          sync.Mutex
	predictions map[string]*Prediction
	nextID      int
}

func newMockReplicateServer() *mockReplicateServer {
	m := &mockReplicateServer{
		predictions: make(map[string]*Prediction),
	}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.handle(w, r)
	}))

	return m
}

func (m *mockReplicateServer) close() {
	m.server.Close()
}

func (m *mockReplicateServer) handle(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/predictions":
		m.handleCreate(w, r)
	case r.Method == http.MethodGet && len(r.URL.Path) > len("/predictions/"):
		m.handleGet(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}
}

func (m *mockReplicateServer) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreatePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	m.nextID++
	id := fmt.Sprintf("pred-%d", m.nextID)
	now := time.Now()

	prediction := &Prediction{
		ID:        id,
		Version:   req.Version,
		Status:    StatusStarting,
		Input:     req.Input,
		CreatedAt: now,
	}

	m.predictions[id] = prediction

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prediction)
}

func (m *mockReplicateServer) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/predictions/"):]
	prediction, ok := m.predictions[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "prediction not found"})
		return
	}

	json.NewEncoder(w).Encode(prediction)
}

func (m *mockReplicateServer) completePrediction(id string, output any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, ok := m.predictions[id]; ok {
		now := time.Now()
		p.Status = StatusSucceeded
		p.Output = output
		p.CompletedAt = &now
	}
}

func (m *mockReplicateServer) failPrediction(id string, errMsg string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p, ok := m.predictions[id]; ok {
		now := time.Now()
		p.Status = StatusFailed
		p.Error = errMsg
		p.CompletedAt = &now
	}
}
