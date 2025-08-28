package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func ParseCfg(fname string, fn Executable) (*Process, error) {
	cfg, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("svr: failed to open %q: %v", fname, err)
	}
	defer func() {
		if err := cfg.Close(); err != nil {
			log.Printf("svr: failed to close %q: %v", fname, err)
		}
	}()

	const (
		programLineMarker = "[program"
		commandLineMarker = "command"
		queueMarker       = "queue"
	)

	var (
		name    string
		process []string
	)

	scanner := bufio.NewScanner(cfg)
	for scanner.Scan() {
		switch line := scanner.Text(); {
		case strings.Contains(line, programLineMarker):
			name = strings.Trim(strings.Split(line, ":")[1], "]")
		case strings.Contains(line, commandLineMarker):
			fullCmd := strings.TrimLeft(strings.Split(line, "=")[1], " ")
			cmdElems := strings.Fields(fullCmd)
			queueName := strings.Split(cmdElems[2], "/")

			process = cmdElems[:2]
			process = append(process, queueName[0]+"/info")

			if len(cmdElems) > 4 && !strings.HasPrefix(cmdElems[3], "--") {
				process = append(process, cmdElems[3])
			}
		}
	}

	if name == "" || len(process) < 3 || !strings.Contains(process[2], queueMarker) {
		return nil, fmt.Errorf("svr: not queue config: %s", cfg.Name())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("svr: failed to scan %q: %w", fname, err)
	}

	return New(name, process, fn), nil
}
