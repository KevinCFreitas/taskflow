FROM golang:1.27-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY main.go ./
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/taskflow .

FROM alpine:3.20

WORKDIR /app
COPY --from=build /out/taskflow /app/taskflow
COPY public ./public
COPY storage ./storage

ENV PORT=3000
EXPOSE 3000
CMD ["/app/taskflow"]
