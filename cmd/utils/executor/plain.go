// Copyright 2022 The Okteto Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package executor

import (
	"bufio"
	"context"
	"os/exec"

	oktetoLog "github.com/okteto/okteto/pkg/log"
)

type plainExecutor struct {
	stdoutScanner *bufio.Scanner
	stderrScanner *bufio.Scanner

	commandContext context.Context
	cancel         context.CancelFunc
}

func newPlainExecutor() *plainExecutor {
	return &plainExecutor{}
}

func (e *plainExecutor) startCommand(cmd *exec.Cmd) error {
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	e.stdoutScanner = bufio.NewScanner(stdoutReader)

	stderrReader, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	e.stderrScanner = bufio.NewScanner(stderrReader)
	return startCommand(cmd)
}

func (e *plainExecutor) display(_ string) {
	e.commandContext, e.cancel = context.WithCancel(context.Background())
	go func() {
		for e.stdoutScanner.Scan() {
			select {
			case <-e.commandContext.Done():
			default:
				line := e.stdoutScanner.Text()
				oktetoLog.Println(line)
				continue
			}
			break
		}
	}()

	go func() {
		for e.stderrScanner.Scan() {
			select {
			case <-e.commandContext.Done():
			default:
				line := e.stderrScanner.Text()
				oktetoLog.Warning(line)
				continue
			}
			break
		}
	}()
}

func (e *plainExecutor) cleanUp(_ error) {
	e.cancel()
	<-e.commandContext.Done()
}