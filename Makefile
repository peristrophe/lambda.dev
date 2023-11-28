AWS_ACCOUNT_ID := xxxxxxxxxxxx
AWS_REGION := ap-northeast-1
AWS_ECR_REGISTRY := $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com

clean-pushed:
	-docker rmi -f $(shell docker images -qf "reference=$(AWS_ECR_REGISTRY)/*:*")
