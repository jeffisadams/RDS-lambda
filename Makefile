
# Ensuring we have some default variable names when they aren't otherwise set
ifeq ($(STACK_NAME),)
STACK_NAME := 'RDS-lambda-test-stack'
endif

ifeq ($(STACK_BUCKET),)
STACK_BUCKET := 'RDS-test-stack-plumbing-bucket'
endif

ifeq ($(DB_SERVICE_USER),)
DB_SERVICE_USER := 'lambda_user'
endif

ifeq ($(DB_ADMIN_USER),)
DB_ADMIN_USER := 'admin'
endif

ifeq ($(DB_PASSWORD),)
DB_PASSWORD := 'testPassword'
endif

ifeq ($(DB_NAME),)
DB_NAME := 'testdatabase'
endif

ifeq ($(DB_NAME),)
DB_TABLE_NAME := 'transactions'
endif

.PHONY: test
test:
	aws cloudformation validate-template --template-body file://template.yaml

.PHONY: clean
clean:
	rm -rf ./dist
	rm -rf template_deploy.yaml

.PHONY: deps
deps: clean
	go get github.com/aws/aws-lambda-go/events
	go get github.com/aws/aws-lambda-go/lambda
	go get github.com/aws/aws-sdk-go/aws/session
	go get github.com/jmoiron/sqlx
	go get github.com/go-sql-driver/mysql

.PHONY: build
build: deps
	GOOS=linux go build -o dist/deploy ./src/deploy.go
	GOOS=linux go build -o dist/main ./src/main.go
	GOOS=linux go build -o dist/test_deploy ./src/test_deploy.go

# Leaving this here as a tidbit of extra info, but it's not 
# .PHONY: api
# api: build
# 	sam local start-api --env-vars env.json

.PHONY: deploy
deploy:
	aws cloudformation package \
		--template-file template.yaml \
		--output-template template_deploy.yaml \
		--s3-bucket $(STACK_BUCKET)

	aws cloudformation deploy \
		--no-fail-on-empty-changeset \
		--template-file template_deploy.yaml \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_IAM

.PHONY: teardown
teardown:
	aws cloudformation delete-stack --stack-name $(STACK_NAME)
	aws cloudformation delete-stack --stack-name $(STACK_NAME)
	clean

.PHONY: init_db
init_db: build

	DB_HOST = $(shell aws rds describe-db-clusters --db-cluster-identifier testdatabase --output text)
	echo $(DB_HOST)
	# ./dist/init


.PHONY: deploy_api
deploy_api:
	aws cloudformation package \
		--template-file template.yaml \
		--output-template api_template_deploy.yaml \
		--s3-bucket $(STACK_BUCKET)

	aws cloudformation deploy \
		--no-fail-on-empty-changeset \
		--template-file api_template_deploy.yaml \
		--stack-name $(STACK_NAME)_api \
		--parameter-overrides "RDSClusterID=$(shell aws rds describe-db-clusters --db-cluster-identifier testdatabase --query 'DBClusters[0].DbClusterResourceId' --output text)" \
		--capabilities CAPABILITY_IAM
