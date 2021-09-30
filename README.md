# Alertmanager Signal Receiver

A Prometheus Alertmanager Webhook Receiver that forwards alerts to a group in [Signal](https://signal.org/).

Heavily based on [prometheus-am-executor](https://github.com/imgix/prometheus-am-executor/).

## Setup

Create a volume for signal-cli.

~~~ bash
docker volume create signal-data
~~~

Start a temporary container with access to signal-cli.

~~~ bash
docker run -it --rm -v signal-data:/app/data --entrypoint bash ghcr.io/dadevel/signal-receiver
~~~

Register a new phone number.

~~~ bash
signal-cli --config ./data --username $YOUR_PHONE_NUMBER register
signal-cli --config ./data --username $YOUR_PHONE_NUMBER verify $CODE_RECEIVED_VIA_SMS
~~~

Create a new group (prints base64-encoded group id on success).

~~~ bash
signal-cli --config ./data --username $YOUR_PHONE_NUMBER updateGroup --name Alerts --description 'Alertmanager notifications' --set-permission-send-messages only-admins --member $SOMEONES_PHONE_NUMBER $ANOTHER_PHONE_NUMBER
~~~

Send a test message.

~~~ bash
signal-cli --config ./data --username $YOUR_PHONE_NUMBER send --group $GROUP_ID_FROM_ABOVE --message 'Hello World!'
~~~

Now that signal-cli is ready to go you can exit the temporary container and finally start the webhook receiver.

~~~ sh
docker run -d -p 9709 -v signal-data:/app/data -e SIGNAL_RECEIVER_PHONE_NUMBER=$YOUR_PHONE_NUMBER -e SIGNAL_RECEIVER_GROUP_ID=$YOUR_GROUP_ID ghcr.io/dadevel/signal-receiver
~~~

Test it.

~~~ sh
curl -i -d @- http://localhost:9709/alert << EOF
{
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
      "startsAt": "2016-04-07T18:08:52.804+02:00",
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
}
EOF
~~~

## Configure

| Environment variable               | Description                                                    |
|------------------------------------|----------------------------------------------------------------|
| `SIGNAL_RECEIVER_PHONE_NUMBER`     | phone number of signal account to send messages from, required |
| `SIGNAL_RECEIVER_GROUP_ID`         | signal group id to send messages to, required                  |
| `SIGNAL_RECEIVER_DATA_DIR`         | storage location used by `signal-cli`, defaults to `./data`    |
| `SIGNAL_RECEIVER_LISTEN_ADDRESS`   | address and port to listen on, defaults to `:9709`             |
| `SIGNAL_RECEIVER_VERBOSE`          | enable verbose logging, off by default                         |
| `SIGNAL_RECEIVER_MESSAGE_TEMPLATE` | go template for messages, see source code for default value    |

Example configurations for Prometheus and Alertmanager can be found in the [examples](./examples) directory.

## Build

Build the binary.

~~~ bash
go build -o ./alertmanager-signal-receiver ./cmd/main.go
~~~

The `Dockerfile` is available under [github.com/dadevel/dockerfiles](https://github.com/dadevel/dockerfiles/tree/main/signal-receiver).
