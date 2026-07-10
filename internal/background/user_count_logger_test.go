package background

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"
)

type userCounterStub struct {
	count func(context.Context) (int64, error)
}

func (s userCounterStub) Count(ctx context.Context) (int64, error) {
	return s.count(ctx)
}

func TestRunUserCountLoggerLogsCountAndStopsOnCancellation(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	called := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		RunUserCountLogger(ctx, logger, userCounterStub{
			count: func(context.Context) (int64, error) {
				called <- struct{}{}
				return 7, nil
			},
		}, 5*time.Millisecond)
		close(done)
	}()

	select {
	case <-called:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("counter was not called")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry["msg"] != "total users" {
		t.Fatalf("msg = %#v, want %q", entry["msg"], "total users")
	}
	if entry["count"] != float64(7) {
		t.Fatalf("count = %#v, want 7", entry["count"])
	}
}

func TestRunUserCountLoggerLogsCountError(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	called := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		RunUserCountLogger(ctx, logger, userCounterStub{
			count: func(context.Context) (int64, error) {
				called <- struct{}{}
				return 0, errors.New("database unavailable")
			},
		}, 5*time.Millisecond)
		close(done)
	}()

	select {
	case <-called:
		cancel()
	case <-time.After(time.Second):
		t.Fatal("counter was not called")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("worker did not stop after cancellation")
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	if entry["msg"] != "failed to count users" {
		t.Fatalf("msg = %#v, want %q", entry["msg"], "failed to count users")
	}
}
