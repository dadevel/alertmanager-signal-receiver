FROM debian:unstable-slim as signal-cli
WORKDIR /build
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update && apt-get install --no-install-recommends -y ca-certificates curl dos2unix
RUN VERSION=$(curl -sSI https://github.com/AsamK/signal-cli/releases/latest | grep -i ^location: | dos2unix | grep -Eo '[0-9.]+$') && \
curl -L https://github.com/AsamK/signal-cli/releases/download/v$VERSION/signal-cli-$VERSION.tar.gz | tar -xzf - && \
mv ./signal-cli-$VERSION/ ./signal-cli/

FROM golang:latest as alertmanager-signal-receiver
WORKDIR /build
COPY . .
RUN go build -o ./alertmanager-signal-receiver ./cmd/main.go
RUN strip ./alertmanager-signal-receiver

FROM debian:unstable-slim
WORKDIR /app
ENV DEBIAN_FRONTEND noninteractive
# without this directory the dpkg install script of openjdk fails
RUN mkdir -p /usr/share/man/man1 && \
   apt-get update && \
   apt-get install --no-install-recommends -y locales openjdk-16-jre-headless && \
   mkdir ./data && \
   chown -R 1000:1000 ./data && \
   sed -i '/en_US.UTF-8/s/^# //g' /etc/locale.gen && \
   locale-gen
COPY --from=signal-cli /build/signal-cli/bin/signal-cli ./bin/
COPY --from=signal-cli /build/signal-cli/lib/ ./lib/
COPY --from=alertmanager-signal-receiver /build/alertmanager-signal-receiver ./bin/
ENV PATH /app/bin:$PATH
ENV LANG en_US.UTF-8
ENV LANGUAGE en_US:en
ENV LC_ALL en_US.UTF-8
USER 1000:1000
ENTRYPOINT ["alertmanager-signal-receiver"]
EXPOSE 9709/tcp
