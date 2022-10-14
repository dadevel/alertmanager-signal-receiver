package main

import (
	"log"
	"os"

	"github.com/dadevel/alertmanager-signal-receiver/message"
	"github.com/dadevel/alertmanager-signal-receiver/signal"
	"github.com/dadevel/alertmanager-signal-receiver/webhook"
)

func main() {
	snd, err := signal.New(
		os.Getenv("SIGNAL_RECEIVER_PHONE_NUMBER"),
		os.Getenv("SIGNAL_RECEIVER_GROUP_ID"),
		os.Getenv("SIGNAL_RECEIVER_DATA_DIR"),
	)
	if err != nil {
		log.Fatal(err)
	}
	go snd.Drain()
	srv := webhook.New(
		os.Getenv("SIGNAL_RECEIVER_LISTEN_ADDRESS"),
		os.Getenv("SIGNAL_RECEIVER_VERBOSE") != "",
		message.New(os.Getenv("SIGNAL_RECEIVER_MESSAGE_TEMPLATE")),
		snd,
	)
	srv.Run()
}
