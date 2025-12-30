package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/mcoder33/qobserver/internal/cmd"
	"github.com/mcoder33/qobserver/internal/model"
	"github.com/mcoder33/qobserver/internal/service"

	"github.com/mcoder33/qobserver/internal/slimtg"
)

const (
	flagConfigDirectoryHelp = "Path to supervisor conf.d directory"
	flagTelegramTokenHelp   = "Telegram bot token"
	flagTelegramChatIdHelp  = "Telegram chat ID"
	flagSleepHelp           = "Sleep between info executing in seconds; use 1s,2s,Ns..."
	flagTTLHelp             = "Command execution ttl; use 1s,2s,Ns..."
	flagMaxWaitHelp         = "Threshold for waiting alert"
	flagMaxDelayHelp        = "Threshold for delayed alert"
	flagVerboseHelp         = "Verbose mode"

	defaultPopulateTime = 5 * time.Minute
)

func main() {
	var (
		flagConfigDir string
		flagTgToken   string
		flagTgChatID  string
		flagSleep     time.Duration
		flagTTL       time.Duration
		flagMaxWait   int
		flagMaxDelay  int
		flagVerbose   bool
	)

	flag.StringVar(&flagConfigDir, "config", "/etc/supervisor/conf.d", flagConfigDirectoryHelp)
	flag.StringVar(&flagTgToken, "tg-token", "", flagTelegramTokenHelp)
	flag.StringVar(&flagTgChatID, "tg-chat-id", "", flagTelegramChatIdHelp)
	flag.DurationVar(&flagSleep, "sleep", 1*time.Second, flagSleepHelp)
	flag.DurationVar(&flagTTL, "ttl", 5*time.Second, flagTTLHelp)
	flag.IntVar(&flagMaxWait, "max-wait", 1000, flagMaxWaitHelp)
	flag.IntVar(&flagMaxDelay, "max-delay", 10000, flagMaxDelayHelp)
	flag.BoolVar(&flagVerbose, "verbose", false, flagVerboseHelp)
	flag.Parse()

	if flagTgToken == "" || flagTgChatID == "" {
		log.Fatal("main: tg-token and tg-chat-id are required")
	}

	pool := cmd.NewPool(func(ctx context.Context, name string, arg ...string) ([]byte, error) {
		return exec.Command(name, arg...).CombinedOutput()
	})

	if err := pool.Populate(flagConfigDir); err != nil {
		log.Fatal("main: no config parsed... Exit!")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		d := ctx.Done()
		t := time.NewTicker(defaultPopulateTime)
		for {
			select {
			case <-d:
				return
			case <-t.C:
				if err := pool.Populate(flagConfigDir); err != nil {
					stop()
					log.Fatal("main: no config parsed... Exit!")
				}
			}
		}
	}()

	tg := slimtg.NewClient(flagTgToken)
	watcher := service.NewWatcher(flagSleep, flagTTL)

	hostname := getHostName()
	for qi := range watcher.Run(ctx, pool) {
		if flagVerbose {
			log.Printf("main: watching %s", qi)
		}

		if qi.Waiting <= flagMaxWait && qi.Delayed <= flagMaxDelay {
			continue
		}

		msg := model.ChatMessage{
			ID:   flagTgChatID,
			Text: "Host: " + hostname + ";\n" + qi.String(),
		}
		if err := tg.Send(msg); err != nil {
			log.Printf("main: sending warning to Telegram: %v", err)
		}
	}
}

func getHostName() (hostname string) {
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("main: failed to get hostname: %v", err)
		hostname = "Unknown"
	}
	return
}
