package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "net/http"
    "os"
    "os/exec"
    "strings"

    "github.com/prometheus/alertmanager/template"
)

var (
    sender string
    group string
    dataDir = "./data"
    address = ":9709"
    logger = log.New(os.Stderr, "", 0)
)

func healthz(res http.ResponseWriter, req *http.Request) {
    io.WriteString(res, "OK\n")
}

func alert(res http.ResponseWriter, req *http.Request) {
    data, err := ioutil.ReadAll(req.Body)
    if err != nil {
        writeError(res, fmt.Errorf("failed to read request body: %w", err), http.StatusBadRequest)
        return
    }
    payload := &template.Data{}
    err = json.Unmarshal(data, payload)
    if err != nil {
        writeError(res, fmt.Errorf("failed to unmarshal payload: %w", err), http.StatusBadRequest)
        return
    }
    var message []string
    for _, alert := range payload.Alerts {
        name := mapGet(alert.Labels, "alertname", "unknown alert")
        instance := mapGet(alert.Labels, "instance", "unknown instance")
        message = append(message, name + " " + alert.Status + " at " + instance)
    }
    err = sendMessage(message)
    if (err != nil) {
        writeError(res, fmt.Errorf("failed to send message: %w", err), http.StatusInternalServerError)
        return
    }
}

func mapGet(mapping map[string]string, key string, defaultValue string) string {
    value, ok := mapping[key]
    if ok {
        return value
    }
    return defaultValue
}

func sendMessage(message []string) error {
    logger.Printf("sending message: %s", strings.Join(message, ", "))
    var buffer bytes.Buffer
    cmd := exec.Command("signal-cli", "--config", dataDir, "--username", sender, "send", "--group", group)
    cmd.Stdin = strings.NewReader(strings.Join(message, "\n"))
    cmd.Stdout = &buffer
    cmd.Stderr = &buffer
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("%v: %w", buffer.String(), err)
    }
    return nil
}

func writeError(res http.ResponseWriter, err error, code int) {
    logger.Println(err)
    http.Error(res, err.Error(), code)
}

func loadEnvironment() {
    sender = os.Getenv("SIGNAL_RECEIVER_PHONE_NUMBER")
    group = os.Getenv("SIGNAL_RECEIVER_GROUP_ID")
    if sender == "" || group == "" {
        logger.Fatal("SIGNAL_RECEIVER_PHONE_NUMBER and/or SIGNAL_RECEIVER_GROUP_ID environment variable not set")
    }
    if os.Getenv("SIGNAL_RECEIVER_DATA_DIR") != "" {
        dataDir = os.Getenv("SIGNAL_RECEIVER_DATA_DIR")
    }
    if os.Getenv("SIGNAL_RECEIVER_LISTEN_ADDRESS") != "" {
        address = os.Getenv("SIGNAL_RECEIVER_LISTEN_ADDRESS")
    }
}

func startServer() {
    http.HandleFunc("/healthz", healthz)
    http.HandleFunc("/alert", alert)
    logger.Printf("listening on %v", address)
    logger.Fatal(http.ListenAndServe(address, nil))
}

func main() {
    loadEnvironment()
    startServer()
}

