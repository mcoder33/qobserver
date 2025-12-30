package cmd

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Pool struct {
	execFn   Executable
	sync     sync.Mutex
	commands map[string]*Process
}

func NewPool(execFn Executable) *Pool {
	return &Pool{
		execFn:   execFn,
		commands: make(map[string]*Process),
	}
}

func (p *Pool) GetAll() map[string]*Process {
	p.sync.Lock()
	defer p.sync.Unlock()

	return maps.Clone(p.commands)
}

func (p *Pool) Populate(cfgDir string) error {
	p.sync.Lock()
	defer p.sync.Unlock()

	files, err := os.ReadDir(cfgDir)
	if err != nil {
		return fmt.Errorf("svr: failed to read %q: %w", cfgDir, err)
	}

	seen := make(map[string]struct{}, len(files))
	const configFileExtension = ".conf"
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), configFileExtension) {
			continue
		}
		seen[file.Name()] = struct{}{}

		if p, ok := p.commands[file.Name()]; ok {
			pinfo, err1 := p.file.Info()
			finfo, err2 := file.Info()
			if err1 == nil && err2 == nil && pinfo.ModTime().Equal(finfo.ModTime()) {
				continue
			}
		}

		fullPath := filepath.Join(cfgDir, file.Name())
		svCfg, err := ParseCfg(fullPath, p.execFn)
		if err != nil {
			log.Printf("svr: Config parse error: %v", err)
			delete(p.commands, file.Name())
			continue
		}
		svCfg.file = file

		p.commands[file.Name()] = svCfg
	}

	for name := range p.commands {
		if _, ok := seen[name]; !ok {
			delete(p.commands, name)
		}
	}

	if len(p.commands) == 0 {
		return fmt.Errorf("svr: no config parsed... Exit")
	}

	return nil
}
