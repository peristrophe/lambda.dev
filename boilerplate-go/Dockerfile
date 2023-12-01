FROM golang:1.20.5-bullseye as builder

COPY *.go /root
COPY go.mod /root
COPY go.sum /root

WORKDIR /root

RUN go mod download
RUN go build
RUN strip lambda_function

# ============================== For ECR PUSH ==============================

FROM public.ecr.aws/lambda/go:1.2023.11.15.20 as deploy

COPY --from=builder /root/lambda_function ${LAMBDA_TASK_ROOT}/lambda_function
COPY config.toml ${LAMBDA_TASK_ROOT}

ENTRYPOINT ["sh", "-c", "${LAMBDA_TASK_ROOT}/lambda_function"]

# ============================== For DEVELOPMENT ==============================

FROM golang:1.20.5-bullseye as development

RUN apt-get update
RUN apt-get install -y --no-install-recommends jq unzip

RUN curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "/tmp/awscliv2.zip"
RUN unzip /tmp/awscliv2.zip -d /tmp/
RUN /tmp/aws/install -i /usr/local/aws-cli -b /usr/local/bin

RUN go install -v golang.org/x/tools/gopls@latest
