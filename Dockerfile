FROM --platform=linux/amd64 golang:1.18
ARG PORT_NUMBER
ARG ENV

ENV PORT=$PORT_NUMBER
ENV APP_ENV=$ENV

WORKDIR /opt
COPY . .
RUN go mod tidy
EXPOSE $PORT
ENTRYPOINT ["go", "run", "server.go"]
