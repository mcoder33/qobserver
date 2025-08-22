package svr

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type watcher struct {
	sleep, ttl time.Duration
}

func NewWatcher(sleep, ttl time.Duration) *watcher {
	return &watcher{sleep: sleep, ttl: ttl}
}

func (o *watcher) Run(ctx context.Context, commands []*Cmd) <-chan *QueueInfo {
	wg := &sync.WaitGroup{}

	out := make(chan *QueueInfo, len(commands))
	ticker := time.NewTicker(o.sleep)

	go func() {
		defer close(out)
		defer ticker.Stop()

	Loop:
		for {
			select {
			case <-ctx.Done():
				break Loop
			case <-ticker.C:
			}

			for _, cmd := range commands {
				wg.Add(1)
				go func(cmd *Cmd) {
					defer wg.Done()
					ctxCmd, cancel := context.WithTimeout(ctx, o.ttl)
					defer cancel()

					qi, err := cmd.Execute(ctxCmd)
					if err != nil {
						if !errors.Is(err, context.Canceled) {
							log.Printf("svr: cancel executing cmd %s: %v", cmd.Name(), err)
						}
						return
					}

					select {
					case <-ctxCmd.Done():
						return
					case out <- qi:
					}
				}(cmd)
			}
			wg.Wait()
		}
	}()

	return out
}
