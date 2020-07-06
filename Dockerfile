FROM gradle:jre14 as signal
WORKDIR /build
RUN git clone --depth 1 https://github.com/AsamK/signal-cli.git .
RUN ./gradlew build && ./gradlew installDist

FROM golang:alpine as receiver
WORKDIR /build
COPY . .
RUN go build

FROM openjdk:14-alpine
WORKDIR /app
COPY --from=signal /build/build/install/signal-cli/bin/ ./bin/
COPY --from=signal /build/build/install/signal-cli/lib/ ./lib/
COPY --from=receiver /build/alertmanager-signal-receiver ./bin/
RUN apk add --no-cache libqrencode libgcc gcompat
RUN mkdir ./data && chown -R nobody:nogroup ./data
USER nobody:nogroup
ENV PATH /app/bin:$PATH
VOLUME /app/data
EXPOSE 9709
ENTRYPOINT ["/app/bin/alertmanager-signal-receiver"]

