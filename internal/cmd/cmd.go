package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mcoder33/qobserver/internal/model"
)

type Executable = func(ctx context.Context, commandName string, arg ...string) ([]byte, error)

type Process struct {
	file    os.DirEntry
	name    string
	command []string
	execFn  Executable
}

func New(name string, command []string, execFn Executable) *Process {
	return &Process{name: name, command: command, execFn: execFn}
}

func (c *Process) Name() string {
	return c.name
}

func (c *Process) Execute(ctx context.Context) (*model.QueueInfo, error) {
	var waiting, delayed, reserved, done int

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("svr: context canceled in cmd %v: %w", c, ctx.Err())
	default:
	}
	out, err := c.execFn(ctx, c.command[0], c.command[1:]...)
	if err != nil {
		return nil, err
	}

	convFunc := func(line string) (int, error) {
		idx := strings.IndexByte(line, ':')
		if idx == -1 || idx+1 >= len(line) {
			return 0, fmt.Errorf("svr: invalid line: %q", line)
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
			return nil, fmt.Errorf("svr: context canceled in cmd %v: %w", c, ctx.Err())
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
			return nil, fmt.Errorf("svr: unexpected cmd response. convFunc fail: %w", err)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("svr: unexpected cmd response. scanner fail: %w", err)
	}

	return &model.QueueInfo{Name: c.Name(), Waiting: waiting, Delayed: delayed, Reserved: reserved, Done: done}, err
}
