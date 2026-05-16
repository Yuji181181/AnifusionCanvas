package domain

import "testing"

func TestJobCanTransitionTo(t *testing.T) {
	tests := []struct {
		name   string
		from   JobStatus
		to     JobStatus
		expect bool
	}{
		{"queued to running", JobStatusQueued, JobStatusRunning, true},
		{"queued to failed", JobStatusQueued, JobStatusFailed, true},
		{"queued to succeeded", JobStatusQueued, JobStatusSucceeded, false},
		{"queued to queued", JobStatusQueued, JobStatusQueued, false},
		{"running to succeeded", JobStatusRunning, JobStatusSucceeded, true},
		{"running to failed", JobStatusRunning, JobStatusFailed, true},
		{"running to running", JobStatusRunning, JobStatusRunning, false},
		{"running to queued", JobStatusRunning, JobStatusQueued, false},
		{"succeeded to anything", JobStatusSucceeded, JobStatusRunning, false},
		{"succeeded to failed", JobStatusSucceeded, JobStatusFailed, false},
		{"failed to anything", JobStatusFailed, JobStatusRunning, false},
		{"failed to succeeded", JobStatusFailed, JobStatusSucceeded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{Status: tt.from}
			got := job.CanTransitionTo(tt.to)
			if got != tt.expect {
				t.Fatalf("expected %v for %s -> %s, got %v", tt.expect, tt.from, tt.to, got)
			}
		})
	}
}
