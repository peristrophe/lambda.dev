LAMBDA_FUNC_NAME := sample-lambda-function-go
AWS_ACCOUNT_ID := xxxxxxxxxxxx
AWS_REGION := ap-northeast-1
AWS_ECR_REGISTRY := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
AWS_ECR_REPOSITORY := lambda-images
PUSH_TO := $(AWS_ECR_REGISTRY)/$(AWS_ECR_REPOSITORY):$(LAMBDA_FUNC_NAME)

define ImportCredentials
$(eval export AWS_ACCESS_KEY_ID=$(shell jq -r '.Credentials.AccessKeyId' .cache/credentials.json 2>/dev/null))
$(eval export AWS_SECRET_ACCESS_KEY=$(shell jq -r '.Credentials.SecretAccessKey' .cache/credentials.json 2>/dev/null))
$(eval export AWS_SESSION_TOKEN=$(shell jq -r '.Credentials.SessionToken' .cache/credentials.json 2>/dev/null))
$(eval export AWS_DEFAULT_REGION=ap-northeast-1)
$(eval export AWS_DEFAULT_OUTPUT=json)
endef

trim:
	go mod tidy

build:
	docker builder prune --force
	docker build -t $(AWS_ECR_REPOSITORY):$(LAMBDA_FUNC_NAME) --target deploy .

run:
	$(ImportCredentials)
	docker run --rm \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_SESSION_TOKEN \
		-e AWS_DEFAULT_REGION \
		-e AWS_DEFAULT_OUTPUT \
		-p 9000:8080 \
		$(AWS_ECR_REPOSITORY):$(LAMBDA_FUNC_NAME)

test:
	curl -d '@request.json' http://localhost:9000/2015-03-31/functions/function/invocations

push:
	$(MAKE) clean
	$(call ImportCredentials)
	aws ecr get-login-password --region $(AWS_REGION) | docker login --username AWS --password-stdin $(AWS_ECR_REGISTRY)
	$(MAKE) build
	docker tag $(AWS_ECR_REPOSITORY):$(LAMBDA_FUNC_NAME) $(PUSH_TO)
	docker push $(PUSH_TO)

clean:
	-docker rmi -f $(shell docker images -qf "reference=$(AWS_ECR_REPOSITORY):$(LAMBDA_FUNC_NAME)")

debug-image:
	docker run --rm -it -v $(shell pwd):/go/src golang:1.20.5-bullseye bash

