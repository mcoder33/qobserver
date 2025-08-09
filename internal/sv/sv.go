package sv

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Executable = func(commandName string, arg ...string) ([]byte, error)

type QueueInfo struct {
	Waiting  int
	Delayed  int
	Reserved int
	Done     int
}

// TODO: make NewCmd cmd with default with exec.Command
type Cmd struct {
	name    string
	command []string
	execFn  Executable
}

func (c *Cmd) Name() string {
	return c.name
}

func (c *Cmd) Execute(ctx context.Context) (*QueueInfo, error) {
	var waiting, delayed, reserved, done int

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	out, err := c.execFn(c.command[0], c.command[1:]...)
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

	return &QueueInfo{waiting, delayed, reserved, done}, err
}

func ParseCfg(fname string, fn Executable) (*Cmd, error) {
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
		case strings.Contains(line, "cmd"):
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

	return &Cmd{name: name, command: cmd, execFn: fn}, nil
}
