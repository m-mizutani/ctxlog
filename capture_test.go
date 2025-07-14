package ctxlog_test

import (
	"testing"

	"github.com/m-mizutani/ctxlog"
)

func TestCapture(t *testing.T) {
	ctx := t.Context()

	// Create capture
	captureCtx, capture := ctxlog.NewCapture(ctx)
	logger := ctxlog.From(captureCtx)

	// Log some messages
	logger.Info("test message 1")
	logger.Info("test message 2")

	// Check captured messages
	messages := capture.Messages()
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
	if messages[0] != "test message 1" {
		t.Errorf("Expected 'test message 1', got '%s'", messages[0])
	}
	if messages[1] != "test message 2" {
		t.Errorf("Expected 'test message 2', got '%s'", messages[1])
	}

	// Check captured records
	records := capture.Records()
	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}
}
