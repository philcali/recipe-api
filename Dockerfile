FROM public.ecr.aws/lambda/go:1

COPY bootstrap ${LAMBDA_TASK_ROOT}

CMD ["bootstrap"]