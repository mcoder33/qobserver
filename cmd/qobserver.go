package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"qobserver/internal/sv"
	"strings"
	"sync"
	"time"
)

var (
	config = flag.String("conf.d", "/etc/sv/conf.d", "Path to sv conf.d directory")
	sleep  = flag.Int("sleep", 1, "Sleep between info executing in seconds")
	//threshold = flag.Int("threshold", 1000, "Threshold for waiting alert")
)

var (
	qiStorage storage
	cmdPool   []*sv.Cmd
	mtx       sync.Mutex
)

type storage map[string]*sv.QueueInfo

func initialize() {
	flag.Parse()
	qiStorage = make(map[string]*sv.QueueInfo)

	fillCmdPool()
}

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
		if !strings.HasSuffix(file.Name(), ".conf") {
			continue
		}
		execFn := func(ctx context.Context, name string, arg ...string) ([]byte, error) {
			return exec.Command(name, arg...).Output()
		}
		if svCfg, err := sv.ParseCfg(file.Name(), execFn); err == nil {
			cmdPool = append(cmdPool, svCfg)
		}
	}
}

// TODO: переписать на каналы передачу конфигов и добавить контекст
func main() {
	initialize()

	//TODO: написать метод для реакции и отправки месседжа в ТГ
	go func() {
		for {
			for queueName, statInfo := range qiStorage {
				log.Printf("%s:\nwaiting:%d\ndelayed:%d\nreserved:%d\ndone:%d\n", queueName, statInfo.Waiting, statInfo.Delayed, statInfo.Reserved, statInfo.Done)
			}
			time.Sleep(1 * time.Second)
		}
	}()
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
						if !errors.Is(context.Canceled, err) {
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
