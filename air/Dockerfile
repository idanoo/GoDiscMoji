FROM public.ecr.aws/docker/library/golang:1.23.4

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY src/go.mod src/go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]