package webhook_test

import (
	"github.com/dadevel/alertmanager-signal-receiver/defaults"
	"github.com/dadevel/alertmanager-signal-receiver/message"
	"github.com/dadevel/alertmanager-signal-receiver/signal"
	"github.com/dadevel/alertmanager-signal-receiver/webhook"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	tmpl := message.New("")
	sig, err := signal.New("+123456789", "test", "")
	if err != nil {
		t.Errorf("err: %v", err)
	}
	want := webhook.Server{
		Address:  defaults.ListenAddress,
		Verbose:  false,
		Sender:   sig,
		Template: tmpl,
	}
	got := webhook.New("", false, tmpl, sig)
	if want != *got {
		t.Errorf("got: %v, want: %v", *got, want)
	}
}

func TestUnmarshalAlerts(t *testing.T) {
	r := strings.NewReader(exampleRequestBody)
	srv := webhook.Server{}
	alerts, err := srv.UnmarshalAlerts(r)
	if err != nil {
		t.Errorf("err: %v", err)
	}
	got := alerts[0].Status
	want := "firing"
	if got != want {
		t.Errorf("got: %v, want: %v", got, want)
	}
}

const exampleRequestBody = `{
  "receiver": "default",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HelloWorld",
        "instance": "localhost:1234",
        "job": "broken",
        "monitor": "world-monitor"
      },
      "annotations": {},
      "startsAt": "2016-04-07T18:08:52Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": ""
    },
    {
      "status": "firing",
      "labels": {
        "alertname": "HelloWorld",
        "instance": "localhost:5678",
        "job": "broken",
        "monitor": "world-monitor"
      },
      "annotations": {},
      "startsAt": "2016-04-07T18:08:52.804+02:00",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": ""
    }
  ],
  "groupLabels": {
    "alertname": "HelloWorld"
  },
  "commonLabels": {
    "alertname": "HelloWorld",
    "job": "broken",
    "monitor": "world-monitor"
  },
  "commonAnnotations": {},
  "externalURL": "http://localhost:9093",
  "version": "4",
  "groupKey": 9777663806026785000
}`
