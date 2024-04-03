FROM golang:1.22.1-alpine3.19 as builder

ARG CGO_ENABLED=0
WORKDIR /app

COPY . .
RUN go mod tidy
RUN go build -o ./deepenc main.go
RUN touch config.yaml

FROM scratch
COPY --from=builder /app/deepenc /bin/deepenc
COPY --from=builder /app/config.yaml /bin/config.yaml


EXPOSE 1234

ENTRYPOINT ["/bin/deepenc"]

CMD ["-config ./config.yaml"]