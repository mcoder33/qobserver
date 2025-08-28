package service

import (
	"context"
	"errors"
	"github.com/mcoder33/qobserver/internal/cmd"
	"github.com/mcoder33/qobserver/internal/model"
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

func (o *watcher) Run(ctx context.Context, commands []*cmd.Process) <-chan *model.QueueInfo {
	wg := &sync.WaitGroup{}

	out := make(chan *model.QueueInfo, len(commands))
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

			for _, process := range commands {
				wg.Add(1)
				go func(process *cmd.Process) {
					defer wg.Done()
					ctxCmd, cancel := context.WithTimeout(ctx, o.ttl)
					defer cancel()

					qi, err := process.Execute(ctxCmd)
					if err != nil {
						if !errors.Is(err, context.Canceled) {
							log.Printf("svr: cancel executing cmd %s: %v", process.Name(), err)
						}
						return
					}

					select {
					case <-ctxCmd.Done():
						return
					case out <- qi:
					}
				}(process)
			}
			wg.Wait()
		}
	}()

	return out
}
