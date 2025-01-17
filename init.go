package server

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/roadrunner-server/errors"
	"go.uber.org/zap"
)

type command struct {
	log         *zap.Logger
	env         map[string]string
	command     []string
	execTimeout time.Duration
}

func newCommand(log *zap.Logger, env map[string]string, cmd []string, execTimeout time.Duration) *command {
	return &command{
		log:         log,
		env:         env,
		command:     cmd,
		execTimeout: execTimeout,
	}
}

func (b *command) start() error {
	const op = errors.Op("server_on_init")
	stopCh := make(chan struct{}, 1)

	cmd := b.createProcess(b.env, b.command)
	timer := time.NewTimer(b.execTimeout)

	err := cmd.Start()
	if err != nil {
		return errors.E(op, err)
	}

	go func() {
		errW := cmd.Wait()
		if errW != nil {
			b.log.Error("process wait", zap.Error(errW))
		}

		stopCh <- struct{}{}
	}()

	select {
	case <-timer.C:
		err = cmd.Process.Kill()
		if err != nil {
			b.log.Error("process killed", zap.Error(err))
		}
		return nil

	case <-stopCh:
		timer.Stop()
		return nil
	}
}

func (b *command) Write(data []byte) (int, error) {
	b.log.Info(string(data))
	return len(data), nil
}

// create command for the process
func (b *command) createProcess(env map[string]string, cmd []string) *exec.Cmd {
	// cmdArgs contain command arguments if the command in form of: php <command> or ls <command> -i -b
	var cmdArgs []string
	var command *exec.Cmd

	// here we may have 2 cases: command declared as a space separated string or as a slice
	switch len(cmd) {
	// command defined as a space separated string
	case 1:
		// we know that the len is 1, so we can safely use the first element
		cmdArgs = append(cmdArgs, strings.Split(cmd[0], " ")...)
	default:
		// we have a slice with a 2 or more elements
		// first element is the command, the rest are arguments
		cmdArgs = cmd
	}

	if len(cmdArgs) == 1 {
		command = exec.Command(cmd[0])
	} else {
		command = exec.Command(cmdArgs[0], cmdArgs[1:]...)
	}

	// set env variables from the config
	if len(env) > 0 {
		for k, v := range env {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", strings.ToUpper(k), os.Expand(v, os.Getenv)))
		}
	}

	// append system envs
	command.Env = append(command.Env, os.Environ()...)
	// redirect stderr and stdout into the Write function of the process.go
	command.Stderr = b
	command.Stdout = b

	return command
}
