package message

import (
	"bytes"
	"io"
	"strings"
	"text/template"

	alertmanager "github.com/prometheus/alertmanager/template"

	"github.com/dadevel/alertmanager-signal-receiver/defaults"
)

type Template struct {
	Input string
}

type Data struct {
	Status      string
	AlertName   string
	Instance    string
	Severity    string
	Summary     string
	Description string
	Labels      map[string]string
	Annotations map[string]string
}

func New(input string) *Template {
	if input == "" {
		input = defaults.MessageTemplate
	}
	return &Template{Input: input}
}

func (self *Template) Render(alert *alertmanager.Alert) (io.Reader, error) {
	msg := self.ConvertAlert(alert)
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
	}
	tmpl, err := template.New("message").Funcs(funcMap).Parse(self.Input)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, msg)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

func (self *Template) ConvertAlert(alert *alertmanager.Alert) *Data {
	msg := Data{
		Status:      alert.Status,
		AlertName:   alert.Labels["alertname"],
		Instance:    alert.Labels["instance"],
		Severity:    alert.Labels["severity"],
		Summary:     alert.Annotations["summary"],
		Description: alert.Annotations["description"],
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
	for key, value := range alert.Labels {
		if key != "alertname" && key != "instance" && key != "severity" {
			msg.Labels[key] = value
		}
	}
	for key, value := range alert.Annotations {
		if key != "summary" && key != "description" {
			msg.Annotations[key] = value
		}
	}
	return &msg
}
