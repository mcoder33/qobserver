package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os/exec"
	"os/signal"
	"qobserver/internal/sv"
	"sync"
	"syscall"
	"time"
)

var (
	configDir = flag.String("config", "/etc/supervisor/conf.d", "Path to sv conf.d directory")
	sleep     = flag.Duration("sleep", 1, "Sleep between info executing in seconds")
	ttl       = flag.Duration("ttl", 5, "Command execution ttl")
	//threshold = flag.Int("threshold", 1000, "Threshold for waiting alert")
)

func main() {
	flag.Parse()
	pool := sv.NewCmdPool(func(ctx context.Context, name string, arg ...string) ([]byte, error) {
		return exec.Command(name, arg...).Output()
	})
	pool.Populate(*configDir)
	if pool.Empty() {
		log.Fatal("No config parsed... Exit!")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for qi := range watch(ctx, *sleep, *ttl, pool.GetAll()) {
		//TODO: переписать на отправку месседжа в ТГ
		log.Printf("%s:\nwaiting:%d\ndelayed:%d\nreserved:%d\ndone:%d\n", qi.Name, qi.Waiting, qi.Delayed, qi.Reserved, qi.Done)
	}
}

func watch(ctx context.Context, sleep, ttl time.Duration, commands []*sv.Cmd) <-chan *sv.QueueInfo {
	out := make(chan *sv.QueueInfo, len(commands))
	wg := &sync.WaitGroup{}
	ticker := time.NewTicker(sleep)

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
				go pushToOutFromCmd(ctx, wg, ttl, cmd, out)
			}
			wg.Wait()
		}
	}()

	return out
}

func pushToOutFromCmd(ctx context.Context, wg *sync.WaitGroup, ttl time.Duration, cmd *sv.Cmd, out chan<- *sv.QueueInfo) {
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
