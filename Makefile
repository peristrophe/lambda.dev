AWS_ACCOUNT_ID := xxxxxxxxxxxx
AWS_REGION := ap-northeast-1
AWS_ECR_REGISTRY := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

MFA_DEVICE_ARN := arn:aws:iam::$(AWS_ACCOUNT_ID):mfa/xxxxxxx


define ImportCredentials
$(eval export AWS_ACCESS_KEY_ID=$(shell jq -r '.Credentials.AccessKeyId' ./credentials.json 2>/dev/null))
$(eval export AWS_SECRET_ACCESS_KEY=$(shell jq -r '.Credentials.SecretAccessKey' ./credentials.json 2>/dev/null))
$(eval export AWS_SESSION_TOKEN=$(shell jq -r '.Credentials.SessionToken' ./credentials.json 2>/dev/null))
$(eval export AWS_DEFAULT_REGION=ap-northeast-1)
$(eval export AWS_DEFAULT_OUTPUT=json)
endef


mfa-%:
	$(eval export AWS_ACCESS_KEY_ID=XXXXXXXXXXXXXXXXXXXX)
	$(eval export AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx)
	$(eval export AWS_DEFAULT_REGION=ap-northeast-1)
	$(eval export AWS_DEFAULT_OUTPUT=json)
	@docker run --rm \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_DEFAULT_REGION \
		-e AWS_DEFAULT_OUTPUT \
		amazon/aws-cli:latest sts get-session-token \
			--serial-number $(MFA_DEVICE_ARN) \
			--token-code ${@:mfa-%=%} | tee ./credentials.json
	find . -type d -name ".cache" -maxdepth 2 -mindepth 2 -exec cp credentials.json '{}' ';'

clean-pushed:
	-docker rmi -f $(shell docker images -qf "reference=$(AWS_ECR_REGISTRY)/*:*")
