FROM gradle:jdk16-openj9 as signal
WORKDIR /build
RUN git clone --single-branch --depth 1 https://github.com/AsamK/signal-cli.git .
RUN gradle --no-daemon installDist

FROM golang:alpine as receiver
WORKDIR /build
COPY . .
RUN go build -o ./alertmanager-signal-receiver ./cmd/main.go

FROM openjdk:16-alpine
WORKDIR /app
COPY --from=signal /build/build/install/signal-cli/bin/signal-cli ./bin/
COPY --from=signal /build/build/install/signal-cli/lib/ ./lib/
COPY --from=receiver /build/alertmanager-signal-receiver ./bin/
RUN apk add --no-cache libgcc gcompat
RUN mkdir ./data && chown -R nobody:nogroup ./data
ENV PATH /app/bin:$PATH
USER nobody:nogroup
ENTRYPOINT ["alertmanager-signal-receiver"]
EXPOSE 9709/tcp
