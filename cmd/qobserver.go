package main

import (
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
	//sleep     = flag.Int("sleep", 0, "Sleep between info executing in seconds")
	//threshold = flag.Int("threshold", 1000, "Threshold for waiting alert")
)

var (
	qiStorage storage
	svCmdPool []*sv.Cmd
	mtx       sync.Mutex
)

type storage map[string]*sv.QueueInfo

func initialize() {
	flag.Parse()
	qiStorage = make(map[string]*sv.QueueInfo)
}

// TODO: переписать на каналы передачу конфигов и добавить контекст
func main() {
	initialize()

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
		if svCfg, err := sv.ParseSvCfg(file.Name(), execFn); err == nil {
			svCmdPool = append(svCmdPool, svCfg)
		}
	}

	var wg sync.WaitGroup
	for _, svCfg := range svCmdPool {
		wg.Add(1)

		go func(group *sync.WaitGroup) {
			defer wg.Done()
			qi, err := svCfg.Execute()
			if err != nil {
				log.Println(err)
				return
			}

			mtx.Lock()
			qiStorage[svCfg.Name()] = qi
			mtx.Unlock()
		}(&wg)
	}

	//TODO: написать метод для реакции и отправки месседжа в ТГ
	go func() {
		for {
			for queueName, statInfo := range qiStorage {
				log.Printf("%s:\nwaiting:%d\ndelayed:%d\nreserved:%d\ndone:%d\n", queueName, statInfo.Waiting, statInfo.Delayed, statInfo.Reserved, statInfo.Done)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
}
