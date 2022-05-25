FROM --platform=linux/amd64 golang:1.18-alpine

ARG PORT_NUMBER
ARG ENV

ENV PORT=$PORT_NUMBER
ENV APP_ENV=$ENV

WORKDIR /app

COPY . .
RUN go mod download


# RUN go run github.com/99designs/gqlgen generate
RUN go build -buildvcs=false -o /chemindulocal-api

EXPOSE $PORT

CMD [ "/chemindulocal-api" ]