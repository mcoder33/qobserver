package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/mcoder33/qobserver/internal/slimtg"
	"github.com/mcoder33/qobserver/internal/svr"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

var (
	configDir = flag.String("config", "/etc/supervisor/conf.d", "Path to supervisor conf.d directory")
	tgToken   = flag.String("tg-token", "", "Telegram bot token")
	tgChatID  = flag.String("tg-chat-id", "", "Telegram chat ID")
	sleep     = flag.Duration("sleep", 1*time.Second, "Sleep between info executing in seconds; use 1s,2s,Ns...")
	ttl       = flag.Duration("ttl", 5*time.Second, "Command execution ttl; use 1s,2s,Ns...")
	maxWait   = flag.Int("max-wait", 1000, "Threshold for waiting alert")
	maxDelay  = flag.Int("max-delay", 10000, "Threshold for delayed alert")
	verbose   = flag.Bool("verbose", false, "Verbose mode")
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

	tg := slimtg.NewClient(*tgToken)
	watcher := svr.NewWatcher(*sleep, *ttl)
	for qi := range watcher.Run(ctx, cmdPool.GetAll()) {
		if *verbose {
			log.Printf("INFO: watching %s", qi)
		}
		if qi.Waiting <= *maxWait && qi.Delayed <= *maxDelay {
			continue
		}

		hostname, err := os.Hostname()
		if err != nil {
			log.Printf("ERROR: failed to get hostname: %v", err)
			hostname = "Unknown"
		}
		msg := slimtg.ChatMessage{
			ID:   *tgChatID,
			Text: "Host: " + hostname + ";\n" + qi.String(),
		}
		err = tg.Send(msg)
		if err != nil {
			log.Printf("Error sending warning to Telegram: %v", err)
		}
	}
}
