FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/devices-api .

FROM alpine:3.20
WORKDIR /data
COPY --from=build /out/devices-api /usr/local/bin/devices-api
EXPOSE 8080
VOLUME /data
ENTRYPOINT ["devices-api"]
