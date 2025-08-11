package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"qobserver/internal/svr"
	"strings"
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
	verbose   = flag.Bool("v", false, "Verbose mode")
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
		if *verbose {
			log.Printf("INFO: watching %s", qi)
		}
		if qi.Waiting <= *threshold && qi.Delayed <= *threshold && qi.Reserved <= *threshold {
			continue
		}
		err := sendWarningToTg(qi)
		if err != nil {
			log.Printf("Error sending warning to Telegram: %v", err)
		}
	}
}

func sendWarningToTg(qi *svr.QueueInfo) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", *tgToken)
	buf := strings.NewReader(fmt.Sprintf(`{"chat_id": "-%s", "text": "%s"}`, *tgChatID, qi))

	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if *verbose {
		log.Printf("Rsponse: %s, body: %s", resp.Status, string(b))
	}

	return nil
}
