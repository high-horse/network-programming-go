// Listing 3-10: A function that pings a network connection at a regular interval ( ping.go)

package main



import (
	"context"
	"fmt"
	"io"
	"time"
)

const defaultPingInterval = 30 * time.Second


func pinger(ctx context.Context, w io.Writer, reset <- chan time.Duration) {
	var interval time.Duration
	select{
		case <- ctx.Done():
			return
		
		case interval = <- reset : // pulled initial interval off reset channel
		default:
	}
	if interval <= 0 {
		interval = defaultPingInterval
	}
	
	timer := time.NewTimer(interval)
	defer func() {
		if timer.Stop() {
			<- timer.C
		}
	}()
	
	for {
		select{
			case <-ctx.Done():
				return
				
			case newInterval := <-reset:
				if !timer.Stop() {
					<-timer.C
				}
				if interval > 0 {
					interval = newInterval
				}
				
			case <-timer.C:
				if _, err := w.Write([]byte("ping")); err != nil {
					return
				}
		}
		_ = timer.Reset(interval)
	}

}


// Listing 3-11: Testing the pinger and resetting its ping timer interval (ping_example_test.go)
func ExamplePinger() {
	ctx, cancel := context.WithCancel(context.Background())
	r, w := io.Pipe() // in lieu of net.Conn
	
	done := make(chan struct{})
	resetTimer := make(chan time.Duration, 1)
	resetTimer <- time.Second // initial ping interval
	
	go func() {
		pinger(ctx, w, resetTimer)
		close(done)
	}()
	
	recievedPing := func(d time.Duration, r io.Reader) {
		if d >= 0 {
			fmt.Printf("reseting timer (%s)\n", d)
			resetTimer <- d
		}
		
		now := time.Now()
		buf := make([]byte, 1024)
		n, err := r.Read(buf)
		if err != nil {
			fmt.Println(err)
		}
		
		fmt.Printf("recieved %q (%s)\n", buf[:n], time.Since(now).Round(100 * time.Millisecond))
	}
	
	for i, v := range []int64{0, 200, 300, 0, -1, -1, -1} {
		fmt.Printf("Run %d\n", i+1)
		recievedPing(time.Duration(v) * time.Millisecond, r)
	}
	
	cancel()
	<-done //ensures the pinger exits after canceling the context
}