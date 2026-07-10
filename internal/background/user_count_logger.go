package background

import (
	"context"
	"log/slog"
	"time"
)

type UserCounter interface {
	Count(ctx context.Context) (int64, error)
}

// RunUserCountLogger logs the current number of users at the configured interval
// until the parent context is cancelled.
func RunUserCountLogger(
	ctx context.Context,
	logger *slog.Logger,
	counter UserCounter,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			countCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			count, err := counter.Count(countCtx)
			cancel()
			if err != nil {
				logger.ErrorContext(ctx, "failed to count users", "error", err)
				continue
			}

			logger.InfoContext(ctx, "total users", "count", count)
		}
	}
}
