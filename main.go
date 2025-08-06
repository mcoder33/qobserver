package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	config = flag.String("conf.d", "/etc/supervisor/conf.d", "Path to supervisor conf.d directory")
	//sleep     = flag.Int("sleep", 0, "Sleep between info executing in seconds")
	//threshold = flag.Int("threshold", 1000, "Threshold for waiting alert")
)

type Executable = func(name string, arg ...string) ([]byte, error)

var (
	qiStorage storage
	svCmdPool []*svCmd
	mtx       sync.Mutex
)

type svCmd struct {
	name    string
	command []string
}

type storage map[string]*queueInfo

type queueInfo struct {
	waiting  int
	delayed  int
	reserved int
	done     int
}

func initialize() {
	flag.Parse()
	qiStorage = make(map[string]*queueInfo)
}

// TODO: переписать на каналы передачу конфигов и добавить контекст
func main() {
	initialize()

	if *config == "" {
		fmt.Printf("Condig directory for supervisor isn't set\n")
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
		if svCfg, err := parseSvCfg(file.Name()); err == nil {
			svCmdPool = append(svCmdPool, svCfg)
		}
	}

	var wg sync.WaitGroup
	for _, svCfg := range svCmdPool {
		wg.Add(1)

		go func(group *sync.WaitGroup) {
			defer wg.Done()
			qi, err := getInfo(svCfg, func(name string, arg ...string) ([]byte, error) {
				return exec.Command(name, arg...).Output()
			})
			if err != nil {
				log.Println(err)
				return
			}

			mtx.Lock()
			qiStorage[svCfg.name] = qi
			mtx.Unlock()
		}(&wg)
	}

	//TODO: написать метод для реакции и отправки месседжа в ТГ
	go func() {
		for {
			for queueName, statInfo := range qiStorage {
				log.Printf("%s:\nwaiting:%d\ndelayed:%d\nreserved:%d\ndone:%d\n", queueName, statInfo.waiting, statInfo.delayed, statInfo.reserved, statInfo.done)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
}

func getInfo(cmd *svCmd, execFn Executable) (*queueInfo, error) {
	var waiting, delayed, reserved, done int

	out, err := execFn(cmd.command[0], cmd.command[1:]...)
	if err != nil {
		return nil, err
	}

	convFunc := func(line string) (int, error) {
		idx := strings.IndexByte(line, ':')
		if idx == -1 || idx+1 >= len(line) {
			return 0, fmt.Errorf("invalid line: %q", line)
		}
		return strconv.Atoi(strings.TrimSpace(line[idx+1:]))
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.Contains(line, "waiting"):
			waiting, err = convFunc(line)
		case strings.Contains(line, "delayed"):
			delayed, err = convFunc(line)
		case strings.Contains(line, "reserved"):
			reserved, err = convFunc(line)
		case strings.Contains(line, "done"):
			done, err = convFunc(line)
		}

		if err != nil {
			return nil, err
		}
	}

	return &queueInfo{waiting, delayed, reserved, done}, err
}

func parseSvCfg(fname string) (*svCmd, error) {
	cfg, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
	}
	defer cfg.Close()

	var (
		name string
		cmd  []string
	)

	scanner := bufio.NewScanner(cfg)
	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.Contains(line, "[program"):
			name = strings.Trim(strings.Split(line, ":")[1], "]")
		case strings.Contains(line, "command"):
			fullCmd := strings.TrimLeft(strings.Split(line, "=")[1], " ")
			cmdElems := strings.Split(fullCmd, " ")
			queueName := strings.Split(cmdElems[2], "/")

			cmd = cmdElems[:2]
			cmd = append(cmd, queueName[0]+"/info")

			if len(cmdElems) > 4 && !strings.HasPrefix(cmdElems[3], "--") {
				cmd = append(cmd, cmdElems[3])
			}
		}
	}

	if name == "" || len(cmd) < 3 || !strings.Contains(cmd[2], "queue") {
		return nil, fmt.Errorf("not queue config: %s", cfg.Name())
	}

	return &svCmd{name: name, command: cmd}, nil
}
