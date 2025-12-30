package cmd

import (
	"fmt"
	"log"
	"maps"
	"os"
	"path"
	"strings"
	"sync"
)

type Pool struct {
	execFn   Executable
	sync     sync.Mutex
	commands map[os.DirEntry]*Process
}

func NewPool(execFn Executable) *Pool {
	return &Pool{execFn: execFn}
}

func (p *Pool) empty() bool {
	return len(p.commands) == 0
}

func (p *Pool) GetAll() map[os.DirEntry]*Process {
	p.sync.Lock()
	defer p.sync.Unlock()

	return maps.Clone(p.commands)
}

func (p *Pool) Populate(cfgDir string) error {
	files, err := os.ReadDir(cfgDir)
	if err != nil {
		return fmt.Errorf("svr: failed to read %q: %w", cfgDir, err)
	}

	cmds := make(map[os.DirEntry]*Process, len(files))
	const configFileExtension = ".conf"
	for _, file := range files {
		if p, ok := p.commands[file]; ok {
			pinfo, _ := p.file.Info()
			finfo, _ := file.Info()
			if pinfo.ModTime().Equal(finfo.ModTime()) {
				continue
			}
		}
		fullPath := path.Join(cfgDir, file.Name())
		if !strings.HasSuffix(fullPath, configFileExtension) {
			continue
		}
		svCfg, err := ParseCfg(fullPath, p.execFn)
		svCfg.file = file
		if err != nil {
			log.Printf("svr: Config parse error: %e", err)
			continue
		}
		cmds[file] = svCfg
	}
	if p.empty() {
		return fmt.Errorf("svr: no config parsed... Exit")
	}

	p.sync.Lock()
	p.commands = maps.Clone(cmds)
	p.sync.Unlock()

	return nil
}
