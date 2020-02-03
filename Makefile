
ENVIRONMENT        ?= prod
PROJECT            =  innovation
STACK_NAME         =  alexa-framework-for-n00bs
AWS_DEFAULT_REGION ?= eu-west-1
ARTIFACTS_BUCKET   = dw-test-deploy-dev
BUCKET_INTENT_FILES= alexa-framework-for-n00bs-intentfiles


sam_package = aws cloudformation package \
                --template-file sam.yaml \
                --output-template-file dist/sam.yaml \
                --s3-bucket $(ARTIFACTS_BUCKET)

sam_deploy = aws cloudformation deploy \
                --template-file dist/sam.yaml \
                --stack-name $(STACK_NAME) \
		--region $(AWS_DEFAULT_REGION) \
                --parameter-overrides \
                        $(shell cat parameters.conf) \
                --capabilities CAPABILITY_IAM \
                --no-fail-on-empty-changeset
sync_intent_files = aws s3 sync intents/ s3://$(BUCKET_INTENT_FILES)

deploy:
	@mkdir -p dist
	# golang
	cd source/; GOOS=linux go build -ldflags="-s -w" -o main && zip deployment.zip main
	# sam
	$(call sam_package)
	$(call sam_deploy)
	$(call sync_intent_files)
	@rm -rf source/main

clean:
	@rm -rf source/main


