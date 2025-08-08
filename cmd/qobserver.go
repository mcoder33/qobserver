package main

import (
	"context"
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
		execFn := func(name string, arg ...string) ([]byte, error) {
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

func observe(cmdPool []*sv.Cmd, ctx context.Context) <-chan *sv.QueueInfo {
	out := make(chan *sv.QueueInfo)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			//TODO утечка горутин при долгом обходе cmdPool !!!
			//TODO тикер нужно останавливать defer ticker.Stop() (прочитать про тикер!!!)
			case <-time.NewTicker(time.Duration(*sleep) * time.Second).C:
				go func() {
					for _, svCfg := range cmdPool {
						select {
						case <-ctx.Done():
							return
						default:
							qi, err := svCfg.Execute()
							//TODO прерывание всего цикла только из за одной команды
							if err != nil {
								log.Println(err)
								return
							}
							select {
							case <-ctx.Done():
								return
							case out <- qi:
							}
						}
					}
				}()
			}
		}
	}()
	return out
}
