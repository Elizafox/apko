// Copyright 2022 Chainguard, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"bufio"
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"
)

//counterfeiter:generate . executorImplementation
type executorImplementation interface {
	Run(cmd *exec.Cmd, logname string, logger *logrus.Entry) error
}

type defaultBuildImplementation struct{}

func monitorPipe(p io.ReadCloser, logger *logrus.Entry, finish chan struct{}) {
	defer p.Close()

	scanner := bufio.NewScanner(p)
	for scanner.Scan() {
		logger.Debugf("%s", scanner.Text())
	}

	finish <- struct{}{}
}

// Run
func (di *defaultBuildImplementation) Run(
	cmd *exec.Cmd, logname string, baseLogger *logrus.Entry,
) error {
	logger := baseLogger.WithFields(logrus.Fields{"cmd": logname})
	logger.Infof("running: %s", cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	stdout_finish := make(chan struct{})
	stderr_finish := make(chan struct{})

	go monitorPipe(stdout, logger, stdout_finish)
	go monitorPipe(stderr, logger, stderr_finish)

	if err := cmd.Wait(); err != nil {
		return err
	}

	<- stdout_finish
	<- stderr_finish

	return nil
}
