package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"qobserver/internal/sv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	config = flag.String("config", "/etc/supervisor/conf.d", "Path to sv conf.d directory")
	sleep  = flag.Int("sleep", 1, "Sleep between info executing in seconds")
	//threshold = flag.Int("threshold", 1000, "Threshold for waiting alert")
)

var cmdPool []*sv.Cmd

func initialize() {
	flag.Parse()
	fillCmdPool()

	if len(cmdPool) == 0 {
		log.Fatal("No config parsed... Exit!")
	}
}

// TODO переписать под функцию чтобы можно было обложить тестированием
func fillCmdPool() {
	if *config == "" {
		fmt.Printf("Condig directory for sv isn't set\n")
		os.Exit(0)
	}

	files, err := os.ReadDir(*config)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		fullPath := path.Join(*config, file.Name())
		if !strings.HasSuffix(fullPath, ".conf") {
			continue
		}
		execFn := func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return exec.Command(name, arg...).Output()
		}
		svCfg, err := sv.ParseCfg(fullPath, execFn)
		if err != nil {
			log.Printf("Config parse error: %e", err)
			continue
		}
		cmdPool = append(cmdPool, svCfg)
	}
}

// TODO: переписать на каналы передачу конфигов и добавить контекст
func main() {
	initialize()

	//TODO: написать метод для реакции и отправки месседжа в ТГ
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	for qi := range observe(ctx, time.Duration(*sleep)*time.Second, cmdPool) {
		log.Printf("%s:\nwaiting:%d\ndelayed:%d\nreserved:%d\ndone:%d\n", qi.Name, qi.Waiting, qi.Delayed, qi.Reserved, qi.Done)
	}
}

func observe(ctx context.Context, sleep time.Duration, cmdPool []*sv.Cmd) <-chan *sv.QueueInfo {
	out := make(chan *sv.QueueInfo, len(cmdPool))
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

			for _, cmd := range cmdPool {
				wg.Add(1)
				go func(cmd *sv.Cmd) {
					ctxCmd, cancel := context.WithTimeout(ctx, 5*time.Second)
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
				}(cmd)
			}
			wg.Wait()
		}
	}()

	return out
}
