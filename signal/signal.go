package signal

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/dadevel/alertmanager-signal-receiver/defaults"
)

type Sender struct {
	PhoneNumber string
	GroupId     string
	DataDir     string
	lock        sync.Mutex
}

func New(phone string, group string, dir string) (*Sender, error) {
	if phone == "" {
		return nil, fmt.Errorf("empty phone number")
	}
	if group == "" {
		return nil, fmt.Errorf("empty group id")
	}
	if dir == "" {
		dir = defaults.DataDir
	}
	return &Sender{
		PhoneNumber: phone,
		GroupId:     group,
		DataDir:     dir,
	}, nil
}

func (self *Sender) Send(msg io.Reader) error {
	cmd := exec.Command("signal-cli", "--config", self.DataDir, "--username", self.PhoneNumber, "send", "--group", self.GroupId)
	cmd.Stdin = msg
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	self.lock.Lock()
	defer self.lock.Unlock()
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("signal-cli: command execution failed: %s: %w", out.String(), err)
	}
	return nil
}
