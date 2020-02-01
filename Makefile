
ENVIRONMENT        ?= prod
PROJECT            =  innovation
STACK_NAME         =  Alexa-framework-for-n00bs
AWS_DEFAULT_REGION ?= eu-west-1

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

deploy:
	@mkdir -p dist
	# golang
	cd source/; GOOS=linux go build -ldflags="-s -w" -o main && zip deployment.zip main
	# sam
	$(call sam_package)
	$(call sam_deploy)
	@rm -rf source/main

clean:
	@rm -rf source/main


