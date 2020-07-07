# Alertmanager Signal Receiver [![CI](https://github.com/dadevel/alertmanager-signal-receiver/workflows/CI/badge.svg?branch=master)](https://github.com/dadevel/alertmanager-signal-receiver/actions) [![Docker Image Version (latest by date)](https://img.shields.io/docker/v/dadevel/alertmanager-signal-receiver?color=blue&logo=docker)](https://hub.docker.com/r/dadevel/alertmanager-signal-receiver)

A Prometheus Alertmanager Webhook Receiver that forwards alerts to a group in [Signal](https://signal.org/).

Heavily based on [prometheus-am-executor](https://github.com/imgix/prometheus-am-executor/).

## Setup

Create a volume for signal-cli.

~~~ bash
docker volume create signal-data
~~~

Start a temporary container with access to signal-cli.

~~~ bash
docker run -it --rm -v signal-data:/app/data --entrypoint /bin/sh dadevel/alertmanager-signal-receiver -i
~~~

A: Register new phone number

Run the following commands inside the container.

~~~ bash
signal-cli --config ./data --username YOUR_PHONE_NUMBER register
signal-cli --config ./data --username YOUR_PHONE_NUMBER verify PIN_RECEIVED_VIA_SMS
~~~

B: Link existing device

Generate a QR-code and scan it with the Signal app on your phone to link a new device to your account.

~~~ bash
apk add --no-cache libqrencode
signal-cli --config ./data link --name alertmanager | tee /dev/stderr | head -n 1 | qrencode -t UTF8
~~~

Either way continue with creating a new group.

~~~ bash
signal-cli --config ./data --username YOUR_PHONE_NUMBER updateGroup --name Alerts --member SOMEONES_PHONE_PHONE --member ANOTHER_PHONE_NUMBER
~~~

Send a test message.

~~~ bash
signal-cli --config ./data --username YOUR_PHONE_NUMBER send --group ID_PRINTED_BY_PREVIOUS_COMMAND --message "Hello World!"
~~~

Now that signal-cli is ready to go you can exit the temporary container and finally start the webhook receiver.

~~~ sh
docker run -d -p 9709 -v signal-data:/app/data -e SIGNAL_RECEIVER_PHONE_NUMBER=YOUR_PHONE_NUMBER -e SIGNAL_RECEIVER_GROUP_ID=YOUR_GROUP_ID dadevel/alertmanager-signal-receiver
~~~

Test it.

~~~ sh
curl --fail --data @- http://localhost:9709/alert << EOF
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

A configuration snippet for alertmanager can be found in [examples/alertmanager.yaml](./examples/alertmanager.yaml).

## Build

~~~ bash
docker build -t dadevel/alertmanager-signal-receiver .
~~~

