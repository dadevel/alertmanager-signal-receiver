# Alertmanager Signal Receiver [![CI](https://github.com/dadevel/alertmanager-signal-receiver/workflows/CI/badge.svg?branch=master)](https://github.com/dadevel/alertmanager-signal-receiver/actions) [![Docker Image Version (latest by date)](https://img.shields.io/docker/v/dadevel/alertmanager-signal-receiver?color=blue&logo=docker)](https://hub.docker.com/r/dadevel/alertmanager-signal-receiver)

A Prometheus Alertmanager Webhook Receiver that forwards alerts to a group in [Signal](https://signal.org/).

Heavily based on [prometheus-am-executor](https://github.com/imgix/prometheus-am-executor/).

## Setup

Create a volume for signal-cli.

~~~ sh
docker volume create signal-data
~~~

Start a temporary container with access to signal-cli.

~~~
docker run -it --rm -v signal-data:/app/data -e SIGNAL_SENDER=YOUR_PHONE_NUMBER --entrypoint /bin/sh dadevel/alertmanager-signal-receiver
~~~

Run the following commands inside the container.

Generate a QR-code and scan it with the Signal app on your phone to link a new device with you phone number.

~~~ sh
signal-cli --config ./data link --name alertmanager | tee /dev/stderr | head -n 1 | qrencode -t UTF8
~~~

Create a new group.

~~~ sh
signal-cli --config ./data --username "$SIGNAL_SENDER" updateGroup --name Alerts --member SOMEONES_PHONE_PHONE --member ANOTHER_PHONE_NUMBER
~~~

Send a test message.

~~~
signal-cli --config ./data --username "$SIGNAL_SENDER" send --group GROUP_ID_FROM_ABOVE --message "Hello World!"
~~~

Now that signal-cli is ready to go you can exit the temporary container and finally start the webhook receiver.

~~~ sh
docker run -d -p 8080 -v signal-data:/app/data -e SENDER=YOUR_PHONE_NUMBER -e GROUP=YOUR_GROUP_ID dadevel/alertmanager-signal-receiver
~~~

Test it.

~~~ sh
curl --fail --data @- http://localhost:8080/alert << EOF
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

## Build

~~~ sh
docker build -t dadevel/alertmanager-signal-receiver .
~~~

