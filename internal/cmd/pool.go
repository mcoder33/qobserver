package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

type pool struct {
	execFn   Executable
	commands []*Process
}

func NewPool(execFn Executable) *pool {
	return &pool{execFn: execFn}
}

func (p *pool) GetAll() []*Process {
	return p.commands
}

func (p *pool) empty() bool {
	return len(p.commands) == 0
}

func (p *pool) Populate(cfgDir string) error {
	files, err := os.ReadDir(cfgDir)
	if err != nil {
		return fmt.Errorf("svr: failed to read %q: %w", cfgDir, err)
	}

	const configFileExtension = ".conf"
	for _, file := range files {
		fullPath := path.Join(cfgDir, file.Name())
		if !strings.HasSuffix(fullPath, configFileExtension) {
			continue
		}
		svCfg, err := ParseCfg(fullPath, p.execFn)
		if err != nil {
			log.Printf("svr: Config parse error: %e", err)
			continue
		}
		p.commands = append(p.commands, svCfg)
	}

	if p.empty() {
		return fmt.Errorf("svr: no config parsed... Exit")
	}
	return nil
}
