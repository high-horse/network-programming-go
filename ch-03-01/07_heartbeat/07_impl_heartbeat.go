package main

import (
	"context"
	"fmt"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second

func main() {
	ExamplePinger()
}

func pinger(ctx context.Context, w io.Writer, reset <-chan time.Duration) {
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

func ExamplePinger() {
	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	r, w := io.Pipe() // Use io.Pipe() instead of a net.Conn for testing

	done := make(chan struct{})
	resetTimer := make(chan time.Duration) // Unbuffered channel to prevent deadlock

	// Send initial ping interval (1 second) to resetTimer channel
	go func() {
		resetTimer <- time.Second
	}()

	// Start a goroutine for the pinger function
	go func() {
		pinger(ctx, w, resetTimer)
		close(done)
	}()

	// Function to simulate receiving a ping and resetting the timer
	receivePing := func(d time.Duration, r io.Reader) {
		if d >= 0 {
			fmt.Printf("resetting timer (%s)\n", d)
			resetTimer <- d // Reset the timer with the new interval
		}

		now := time.Now()
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Printf("received %q (%s)\n", buf[:n], time.Since(now).Round(100*time.Millisecond))
	}

	// Simulate different intervals and call receivePing for each
	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d:\n", i+1)
		receivePing(time.Duration(v)*time.Millisecond, r)
	}

	// Cancel the context after all pings have been received
	cancel()

	// Ensure the pinger exits after canceling the context
	<-done
}
