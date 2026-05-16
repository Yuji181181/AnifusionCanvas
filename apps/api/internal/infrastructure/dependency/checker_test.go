package dependency

import (
	"context"
	"testing"

	"github.com/haseg/anifusion-canvas/apps/api/internal/config"
)

func TestCheckerSkipsUnsetExternalServices(t *testing.T) {
	checker := NewChecker(config.Config{})

	db := checker.CheckDatabase(context.Background())
	if db.Status != StatusSkipped {
		t.Fatalf("expected database check to be skipped, got %s", db.Status)
	}

	replicate := checker.CheckReplicate(context.Background())
	if replicate.Status != StatusSkipped {
		t.Fatalf("expected replicate check to be skipped, got %s", replicate.Status)
	}

	r2 := checker.CheckR2(context.Background())
	if r2.Status != StatusSkipped {
		t.Fatalf("expected r2 check to be skipped, got %s", r2.Status)
	}
}
