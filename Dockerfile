FROM public.ecr.aws/lambda/go:1

COPY app ${LAMBDA_TASK_ROOT}

CMD ["app"]