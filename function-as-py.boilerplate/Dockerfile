FROM public.ecr.aws/lambda/python:3.11.2023.11.10.19-x86_64 as deploy

COPY requirements.txt ${LAMBDA_TASK_ROOT}
COPY main.py ${LAMBDA_TASK_ROOT}
COPY config.toml ${LAMBDA_TASK_ROOT}
COPY utils ${LAMBDA_TASK_ROOT}/utils

RUN pip install -r requirements.txt

CMD ["main.lambda_handler"]

FROM deploy as development

RUN yum install -y jq aws-cli
