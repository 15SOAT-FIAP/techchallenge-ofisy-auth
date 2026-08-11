# ===== BUILD =====
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -tags lambda.norpc -o bootstrap ./cmd/lambda

# ===== RUNTIME =====
FROM public.ecr.aws/lambda/provided:al2023-arm64

COPY --from=build /app/bootstrap ${LAMBDA_TASK_ROOT}

CMD ["bootstrap"]