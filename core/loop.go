package core

// TODO: rip readline out and replace it with my own implementation

import (
	"errors"
	"fmt"
	"gopnzr/core/lang"
	a "gopnzr/core/shell/args"
	"gopnzr/core/shell/complete"
	"gopnzr/core/shell/config"
	"gopnzr/core/shell/env"
	"gopnzr/core/shell/prompt"
	"gopnzr/core/shell/system"
	"gopnzr/core/state"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chzyer/readline"
)

// main entry point for the shell
// performs the following actions:
//
// 1. computes the value for $PWD
//
// 2. registers notifier for syscalls, such as SIGINT, SIGTERM, etc
//
// 3. computation of at startup known prompt placeholders
//
// 4. starting a go routine for signal handling
//
// 5. main loop
//   - computing the prompt
//   - waiting for input
//   - waiting for input
//   - exiting on EOF (Ctrl+D)
func Shell() {
	args := a.Get()

	if args.Version {
		fmt.Println(state.VERSION, state.VERSION_SUFFIX)
		return
	}

	log.SetFlags(0)
	env.SetEnv("PWD", system.Getwd())
	exe, _ := os.Executable()
	env.SetEnv("SHELL", exe)

	// INFO: this captures interrupts, therefore not killing the shell when
	// terminating commands
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	err := prompt.PreComputePlaceholders()
	if err != nil {
		log.Printf("error while computing prompt placeholders: %s\n", err)
	}

	config.Load(&args)

	if args.Command != "" {
		run(args.Command, &args)
		return
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:       prompt.ComputePrompt(),
		AutoComplete: complete.BuildCompleter(),
	})

	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		rl.SetPrompt(prompt.ComputePrompt())
		input, err := rl.Readline()
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("bye bye, see you :^)")
				os.Exit(0)
			} else if errors.Is(err, readline.ErrInterrupt) {
				continue
			}

			fmt.Fprintln(os.Stderr, err)
		}

		if input == "" {
			continue
		}
		run(input, &args)
	}
}

func run(input string, args *a.Arguments) {
	defer func() {
		if args.Debug {
			return
		}
		if err := recover(); err != nil {
			log.Print(err)
		}
	}()
	lang.Compile(input, args)
}
