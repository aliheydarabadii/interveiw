FROM golang:1.22-bookworm AS build

WORKDIR /src

COPY docker/modbus-simulator/main.go /src/main.go

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
	-trimpath \
	-ldflags="-s -w" \
	-o /out/modbus-simulator \
	./main.go

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/modbus-simulator /app/modbus-simulator

ENTRYPOINT ["/app/modbus-simulator"]
