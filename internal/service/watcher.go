package service

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/mcoder33/qobserver/internal/cmd"
	"github.com/mcoder33/qobserver/internal/model"
)

type Watcher struct {
	sleep, ttl time.Duration
}

func NewWatcher(sleep, ttl time.Duration) *Watcher {
	return &Watcher{sleep: sleep, ttl: ttl}
}

func (o *Watcher) Run(ctx context.Context, pool *cmd.Pool) <-chan *model.QueueInfo {
	var wg sync.WaitGroup

	out := make(chan *model.QueueInfo)
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

			processes := pool.GetAll()
			wg.Add(len(processes))
			for _, process := range processes {
				go func(process *cmd.Process) {
					defer wg.Done()
					ctxCmd, cancel := context.WithTimeout(ctx, o.ttl)
					defer cancel()

					qi, err := process.Execute(ctxCmd)
					if err != nil {
						if !errors.Is(err, context.Canceled) {
							log.Printf("svr: cancel executing cmd %s: %v", process.Name(), err)
						}
						if errors.As(err, &cmd.ExecError{}) {
							if err := pool.Populate(); err != nil {
								log.Printf("svr: error processing cmd %s: %v", process.Name(), err)
							}
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
