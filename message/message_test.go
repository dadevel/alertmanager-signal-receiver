package message_test

import (
	"github.com/dadevel/alertmanager-signal-receiver/defaults"
	"github.com/dadevel/alertmanager-signal-receiver/message"
	alertmanager "github.com/prometheus/alertmanager/template"
	"io/ioutil"
	"testing"
)

func TestNewWithInput(t *testing.T) {
	want := message.Template{Input: "test"}
	got := *message.New("test")
	if want != got {
		t.Errorf("want: %v, got: %v", want, got)
	}
}

func TestNewWithDefaults(t *testing.T) {
	want := message.Template{Input: defaults.MessageTemplate}
	got := *message.New("")
	if want != got {
		t.Errorf("want: %v, got: %v", want, got)
	}
}

func TestConvertAlert(t *testing.T) {
	alert := alertmanager.Alert{
		Status: "firing",
		Labels: map[string]string{
			"alertname": "Test",
		},
		Annotations: map[string]string{
			"summary": "...",
		},
	}
	want := message.Data{Status: "firing", AlertName: "Test", Summary: "..."}
	tmpl := message.Template{Input: ""}
	got := *tmpl.ConvertAlert(&alert)
	if want.Status != got.Status {
		t.Errorf("want: %v, got: %v", want.Status, got.Status)
	}
	if want.AlertName != got.AlertName {
		t.Errorf("want: %v, got: %v", want.AlertName, got.AlertName)
	}
	if want.Summary != got.Summary {
		t.Errorf("want: %v, got: %v", want.Summary, got.Summary)
	}
	if len(want.Labels) != len(got.Labels) {
		t.Errorf("want: %v, got: %v", len(want.Labels), len(got.Labels))
	}
}

func TestConvertEmptyAlert(t *testing.T) {
	alert := alertmanager.Alert{}
	tmpl := message.Template{Input: ""}
	got := tmpl.ConvertAlert(&alert).Severity
	want := alert.Labels["severity"]
	if want != got {
		t.Errorf("want: %v, got: %v", want, got)
	}
}

func TestRender(t *testing.T) {
	alert := alertmanager.Alert{
		Status: "firing",
		Labels: map[string]string{
			"alertname": "Test",
		},
		Annotations: map[string]string{
			"summary": "...",
		},
	}
	want := "FIRING"
	tmpl := message.Template{Input: "{{ .Status | ToUpper }}"}
	r, err := tmpl.Render(&alert)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	txt, err := ioutil.ReadAll(r)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	got := string(txt)
	if want != got {
		t.Errorf("want: %v, got: %v", want, got)
	}
}
