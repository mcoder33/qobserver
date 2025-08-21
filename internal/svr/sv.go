package svr

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

type Executable = func(ctx context.Context, commandName string, arg ...string) ([]byte, error)

type QueueInfo struct {
	Name     string
	Waiting  int
	Delayed  int
	Reserved int
	Done     int
}

func (i QueueInfo) String() string {
	return fmt.Sprintf("QueueName: %s\nWaiting: %d\nDelayed: %d;\nReserved: %d;\nDone: %d;", i.Name, i.Waiting, i.Delayed, i.Reserved, i.Done)
}

type Cmd struct {
	name    string
	command []string
	execFn  Executable
}

func NewCmd(name string, command []string, execFn Executable) *Cmd {
	return &Cmd{name: name, command: command, execFn: execFn}
}

func (c *Cmd) Name() string {
	return c.name
}

func (c *Cmd) Execute(ctx context.Context) (*QueueInfo, error) {
	var waiting, delayed, reserved, done int

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("ERROR: context canceled in cmd %v: %w", c, ctx.Err())
	default:
	}
	out, err := c.execFn(ctx, c.command[0], c.command[1:]...)
	if err != nil {
		return nil, err
	}

	convFunc := func(line string) (int, error) {
		idx := strings.IndexByte(line, ':')
		if idx == -1 || idx+1 >= len(line) {
			return 0, fmt.Errorf("ERROR: invalid line: %q", line)
		}
		return strconv.Atoi(strings.TrimSpace(line[idx+1:]))
	}

	const (
		waitingHeader  = "waiting"
		delayedHeader  = "delayed"
		reservedHeader = "reserved"
		doneHeader     = "done"
	)

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ERROR: context canceled in cmd %v: %w", c, ctx.Err())
		default:
		}

		switch line := scanner.Text(); {
		case strings.Contains(line, waitingHeader):
			waiting, err = convFunc(line)
		case strings.Contains(line, delayedHeader):
			delayed, err = convFunc(line)
		case strings.Contains(line, reservedHeader):
			reserved, err = convFunc(line)
		case strings.Contains(line, doneHeader):
			done, err = convFunc(line)
		}

		if err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return &QueueInfo{c.Name(), waiting, delayed, reserved, done}, err
}

func ParseCfg(fname string, fn Executable) (*Cmd, error) {
	cfg, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to open %q: %v", fname, err)
	}
	defer func() {
		if err := cfg.Close(); err != nil {
			log.Printf("ERROR: failed to close %q: %v", fname, err)
		}
	}()

	const (
		programLineMarker = "[program"
		commandLineMarker = "command"
		queueMarker       = "queue"
	)

	var (
		name string
		cmd  []string
	)

	scanner := bufio.NewScanner(cfg)
	for scanner.Scan() {
		switch line := scanner.Text(); {
		case strings.Contains(line, programLineMarker):
			name = strings.Trim(strings.Split(line, ":")[1], "]")
		case strings.Contains(line, commandLineMarker):
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

	if name == "" || len(cmd) < 3 || !strings.Contains(cmd[2], queueMarker) {
		return nil, fmt.Errorf("ERROR: not queue config: %s", cfg.Name())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ERROR: failed to scan %q: %w", fname, err)
	}

	return NewCmd(name, cmd, fn), nil
}

type cmdPool struct {
	execFn   Executable
	commands []*Cmd
}

func NewCmdPool(execFn Executable) *cmdPool {
	return &cmdPool{execFn: execFn}
}

func (p *cmdPool) Empty() bool {
	return len(p.commands) == 0
}

func (p *cmdPool) GetAll() []*Cmd {
	return p.commands
}

func (p *cmdPool) Populate(cfgDir string) {
	files, err := os.ReadDir(cfgDir)
	if err != nil {
		log.Fatal(err)
	}

	const configFileExtension = ".conf"
	for _, file := range files {
		fullPath := path.Join(cfgDir, file.Name())
		if !strings.HasSuffix(fullPath, configFileExtension) {
			continue
		}
		svCfg, err := ParseCfg(fullPath, p.execFn)
		if err != nil {
			log.Printf("ERROR: Config parse error: %e", err)
			continue
		}
		p.commands = append(p.commands, svCfg)
	}
}
