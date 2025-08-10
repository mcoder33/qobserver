package sv

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

type observer struct {
	sleep, ttl time.Duration
}

func NewObserver(sleep, ttl time.Duration) *observer {
	return &observer{sleep: sleep, ttl: ttl}
}

func (o *observer) Run(ctx context.Context, commands []*Cmd) <-chan *QueueInfo {
	out := make(chan *QueueInfo, len(commands))
	wg := &sync.WaitGroup{}
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
				go pushToOutFromCmd(ctx, wg, o.ttl, cmd, out)
			}
			wg.Wait()
		}
	}()

	return out
}

func pushToOutFromCmd(ctx context.Context, wg *sync.WaitGroup, ttl time.Duration, cmd *Cmd, out chan<- *QueueInfo) {
	ctxCmd, cancel := context.WithTimeout(ctx, ttl)
	defer wg.Done()
	defer cancel()

	qi, err := cmd.Execute(ctxCmd)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			log.Printf("Error executing cmd %s: %v", cmd.Name(), err)
		}
		return
	}
	if qi == nil {
		return
	}

	select {
	case <-ctxCmd.Done():
		return
	case out <- qi:
	}
}
