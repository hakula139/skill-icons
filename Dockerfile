FROM golang:latest AS builder

WORKDIR /app
COPY api/ /app/api/
COPY cmd/ /app/cmd/
COPY assets/ /app/assets/
COPY build.go go.mod go.sum /app/

RUN go run build.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -ldflags="-s -w" -o skillicons ./cmd/server

FROM scratch

COPY --from=builder /app/skillicons /skillicons

EXPOSE 8080

ENTRYPOINT ["/skillicons"]
