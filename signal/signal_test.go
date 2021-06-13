package signal_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
	//alertmanager "github.com/prometheus/alertmanager/template"
	"github.com/dadevel/alertmanager-signal-receiver/defaults"
	"github.com/dadevel/alertmanager-signal-receiver/signal"
)

func TestNewBadArgs(t *testing.T) {
	_, err := signal.New("", "", "")
	if err == nil {
		t.Errorf("want: err, got: nil")
	}
}

func TestNewGoodArgs(t *testing.T) {
	want := signal.Sender{
		PhoneNumber: "+12345678",
		GroupId:     "test",
		DataDir:     defaults.DataDir,
	}
	got, err := signal.New(want.PhoneNumber, want.GroupId, "")
	if err != nil {
		t.Errorf("err: %v", err)
	}
	if want != *got {
		t.Errorf("want: %v, got: %v", want, *got)
	}
}

func TestSend(t *testing.T) {
	tmp, err := ioutil.TempDir("", "go-test-")
	if err != nil {
		t.Errorf("err: %v", err)
	}
	fmt.Println(tmp)
	defer os.RemoveAll(tmp)
	out := path.Join(tmp, "out.txt")
	script := fmt.Sprintf("#!/bin/sh\necho -n $* > %s\n", out)
	err = ioutil.WriteFile(path.Join(tmp, "signal-cli"), []byte(script), 0755)
	os.Setenv("PATH", tmp)
	s, err := signal.New("+123456789", "test", "/data")
	if err != nil {
		t.Errorf("err: %v", err)
	}
	r := strings.NewReader("Hello World!")
	err = s.Send(r)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	txt, err := ioutil.ReadFile(out)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	got := string(txt)
	want := "--config /data --username +123456789 send --group test"
	fmt.Println(got)
	fmt.Println(want)
	if len(want) != len(got) {
		t.Errorf("want: %v, got: %v", len(want), len(got))
	}
	if want != got {
		t.Errorf("want: %v, got: %v", want, got)
	}
}
