package main

import (
	"context"
	"io"
	"time"
)
const defaultPingInterval = 30 * time.Second


func Pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
	var interval time.Duration

	// Initial select to either cancel or set the interval from the reset channel
	select {
	case <-ctx.Done():
		return
	case interval = <-reset: // pulled initial interval off reset channel
	default:
	}

	// Use default interval if none is provided or invalid
	if interval <= 0 {
		interval = defaultPingInterval
	}

	// Create a new timer with the initial interval
	timer := time.NewTimer(interval)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
	}()

	// Main loop to handle context cancellation, reset of the interval, and pinging
	for {
		select {
		case <-ctx.Done(): // If the context is done, return from the function
			return

		case newInterval := <-reset: // Update interval based on reset signal
			if !timer.Stop() {
				<-timer.C
			}
			if newInterval > 0 {
				interval = newInterval
			}

		case <-timer.C: // When the timer expires, send a "ping"
			if _, err := w.Write([]byte("ping")); err != nil {
				// Handle the error and return if there's an issue writing
				return
			}
		}

		// Reset the timer with the updated interval
		_ = timer.Reset(interval)
	}
}