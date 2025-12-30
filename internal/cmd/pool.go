package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"sync"
)

type Pool struct {
	execFn   Executable
	sync     sync.Mutex
	commands []*Process
}

func NewPool(execFn Executable) *Pool {
	return &Pool{execFn: execFn}
}

func (p *Pool) empty() bool {
	return len(p.commands) == 0
}

func (p *Pool) GetAll() []*Process {
	p.sync.Lock()
	defer p.sync.Unlock()

	var r []*Process
	return append(r, p.commands...)
}

func (p *Pool) Populate(cfgDir string) error {
	files, err := os.ReadDir(cfgDir)
	if err != nil {
		return fmt.Errorf("svr: failed to read %q: %w", cfgDir, err)
	}

	cmds := make([]*Process, len(files))
	const configFileExtension = ".conf"
	for i, file := range files {
		fullPath := path.Join(cfgDir, file.Name())
		if !strings.HasSuffix(fullPath, configFileExtension) {
			continue
		}
		svCfg, err := ParseCfg(fullPath, p.execFn)
		if err != nil {
			log.Printf("svr: Config parse error: %e", err)
			continue
		}
		cmds[i] = svCfg
	}
	if p.empty() {
		return fmt.Errorf("svr: no config parsed... Exit")
	}

	p.sync.Lock()
	p.commands = cmds
	p.sync.Unlock()

	return nil
}
