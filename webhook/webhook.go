package webhook

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	alertmanager "github.com/prometheus/alertmanager/template"

	"github.com/dadevel/alertmanager-signal-receiver/defaults"
	"github.com/dadevel/alertmanager-signal-receiver/message"
	"github.com/dadevel/alertmanager-signal-receiver/signal"
)

var logger = log.New(os.Stderr, "server", 0)

type Server struct {
	Address  string
	Verbose  bool
	Sender   *signal.Sender
	Template *message.Template
}

func New(address string, verbose bool, t *message.Template, s *signal.Sender) *Server {
	srv := Server{
		Address:  address,
		Verbose:  verbose,
		Sender:   s,
		Template: t,
	}
	if srv.Address == "" {
		srv.Address = defaults.ListenAddress
	}
	return &srv
}

func (srv *Server) UnmarshalAlerts(r io.Reader) ([]alertmanager.Alert, error) {
	body, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	data := &alertmanager.Data{}
	err = json.Unmarshal(body, data)
	if srv.Verbose {
		logger.Printf("webhook triggered: %#v", data)
	}
	return data.Alerts, err
}

// untested
func (srv *Server) SendAlert(alert *alertmanager.Alert) error {
	msg, err := srv.Template.Render(alert)
	if err != nil {
		return err
	}
	err = srv.Sender.Send(msg)
	return err
}

// untested
func (srv *Server) WriteError(res http.ResponseWriter, err error, code int) {
	logger.Printf("request handler failed: %v", err)
	http.Error(res, err.Error(), code)
}

// untested
func (srv *Server) HandleHealth(res http.ResponseWriter, req *http.Request) {
	_, err := io.WriteString(res, "OK\n")
	if err != nil {
		srv.WriteError(res, err, http.StatusInternalServerError)
	}
}

// untested
func (srv *Server) HandleAlert(res http.ResponseWriter, req *http.Request) {
	alerts, err := srv.UnmarshalAlerts(req.Body)
	if err != nil {
		srv.WriteError(res, err, http.StatusBadRequest)
		return
	}
	for _, alert := range alerts {
		err := srv.SendAlert(&alert)
		if err != nil {
			srv.WriteError(res, err, http.StatusInternalServerError)
			return
		}
	}
}

// untested
func (srv *Server) Run() {
	http.HandleFunc("/healthz", srv.HandleHealth)
	http.HandleFunc("/alert", srv.HandleAlert)
	logger.Printf("listening on %s", srv.Address)
	logger.Fatal(http.ListenAndServe(srv.Address, nil))
}
