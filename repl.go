package main

import (
	"errors"
	"fmt"

	"github.com/chzyer/readline"
	"github.com/illbjorn/echo"
	"github.com/illbjorn/vsx/argv"
	"github.com/illbjorn/vsx/gallery"
)

var (
	pc        = readline.PcItem
	completer = readline.NewPrefixCompleter(
		// query
		pc("query", pc("--name")),
		// download
		pc("download", pc("--name")),
		// install
		pc("install", pc("--name")),
		// backup
		pc("backup", pc("--to")),
	)
)

func EnterREPL(g gallery.Gallery, cfg *Config) error {
	// Init the readline instance
	l, err := readline.NewEx(&readline.Config{
		Prompt:       "\033[31m»\033[0m ",
		HistoryFile:  cfg.HistFilePath,
		AutoComplete: completer,
	})
	if err != nil {
		return fmt.Errorf("failed to init readline: %w", err)
	}

	for {
		// Wait for a command...
		v, err := l.Readline()
		if err != nil {
			if errors.Is(err, readline.ErrInterrupt) {
				return nil
			}
			return fmt.Errorf("failed to read REPL input: %w", err)
		}

		// Tokenize
		tokens := argv.Tokenize(v)

		// Parse
		cmd, err := argv.Parse(tokens)
		if err != nil {
			return fmt.Errorf("failed to parse input: %w", err)
		}

		// Exec
		err = RunCMD(g, cfg, cmd)
		if err != nil {
			echo.Errorf("Failed [%s]: %s.", cmd.Name, err)
		}
	}
}
