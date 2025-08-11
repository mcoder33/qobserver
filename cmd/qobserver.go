package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"qobserver/internal/svr"
	"syscall"
	"time"
)

var (
	configDir = flag.String("config", "/etc/supervisor/conf.d", "Path to supervisor conf.d directory")
	tgToken   = flag.String("tg-token", "", "Telegram bot token")
	tgChatID  = flag.String("tg-chat-id", "", "Telegram chat ID")
	sleep     = flag.Duration("sleep", 1*time.Second, "Sleep between info executing in seconds; use 1s,2s,Ns...")
	ttl       = flag.Duration("ttl", 5*time.Second, "Command execution ttl; use 1s,2s,Ns...")
	threshold = flag.Int("threshold", 1000, "Threshold for waiting alert")
)

func main() {
	flag.Parse()

	if *tgToken == "" || *tgChatID == "" {
		fmt.Println("tg-token and tg-chat-id are required")
		flag.Usage()
		os.Exit(1)
	}

	cmdPool := svr.NewCmdPool(func(ctx context.Context, name string, arg ...string) ([]byte, error) {
		return exec.Command(name, arg...).Output()
	})
	cmdPool.Populate(*configDir)
	if cmdPool.Empty() {
		log.Fatal("No config parsed... Exit!")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	watcher := svr.NewWatcher(*sleep, *ttl)
	for qi := range watcher.Run(ctx, cmdPool.GetAll()) {
		//TODO: переписать на отправку месседжа в ТГ
		log.Printf("%s:\nwaiting:%d\ndelayed:%d\nreserved:%d\ndone:%d\n", qi.Name, qi.Waiting, qi.Delayed, qi.Reserved, qi.Done)
	}
}
