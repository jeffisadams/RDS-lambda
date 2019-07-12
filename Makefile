DB_NAME := testdatabase

# Ensuring we have some default variable names when they aren't otherwise set
ifeq ($(STACK_NAME),)
STACK_NAME := RDS-lambda-test-stack
endif

ifeq ($(USER_IP),)
USER_IP := 108.56.183.61
endif

ifeq ($(STACK_BUCKET),)
STACK_BUCKET := RDS-test-stack-plumbing-bucket
endif

ifeq ($(DB_SERVICE_USER),)
DB_SERVICE_USER := lambda
endif

ifeq ($(DB_ADMIN_USER),)
DB_ADMIN_USER := admin
endif

ifeq ($(DB_PASSWORD),)
DB_PASSWORD := testing1
endif


ifeq ($(DB_TABLE_NAME),)
DB_TABLE_NAME := transactions
endif

.PHONY: test
test:
	aws cloudformation validate-template --template-body file://template.yaml
	aws cloudformation validate-template --template-body file://api_template.yaml

.PHONY: clean
clean:
	rm -rf ./dist
	rm -rf template_deploy.yaml
	rm -rf api_template_deploy.yaml

.PHONY: deps
deps: clean
	go get github.com/aws/aws-lambda-go/events
	go get github.com/aws/aws-lambda-go/lambda
	go get github.com/aws/aws-sdk-go/aws/session
	go get github.com/jmoiron/sqlx
	go get github.com/go-sql-driver/mysql

.PHONY: build
build: deps
	go build -o dist/init ./src/init.go
	GOOS=linux go build -o dist/main ./src/main.go

.PHONY: deploy_rds
deploy_rds:
	aws cloudformation package \
		--template-file template.yaml \
		--output-template template_deploy.yaml \
		--s3-bucket $(STACK_BUCKET)

	aws cloudformation deploy \
		--no-fail-on-empty-changeset \
		--template-file template_deploy.yaml \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_IAM \
		--parameter-overrides "DatabaseName=$(DB_NAME)" "DatabaseUserPassword=$(DB_PASSWORD)" "UserIP=$(USER_IP)"

.PHONY: init_db
init_db:
	./dist/init $(DB_ADMIN_USER) $(DB_PASSWORD) $(shell aws rds describe-db-clusters --db-cluster-identifier testdatabase --query 'DBClusters[0].Endpoint' --output text) $(DB_NAME) $(DB_TABLE_NAME) $(DB_SERVICE_USER)

.PHONY: deploy_api
deploy_api:
	aws cloudformation package \
		--template-file api_template.yaml \
		--output-template api_template_deploy.yaml \
		--s3-bucket $(STACK_BUCKET)

	aws cloudformation deploy \
		--no-fail-on-empty-changeset \
		--template-file api_template_deploy.yaml \
		--stack-name $(STACK_NAME)-api \
		--parameter-overrides "DatabaseName=$(DB_NAME)" "DatabaseTableName=$(DB_TABLE_NAME)" "DatabaseUserName=$(DB_SERVICE_USER)" "RDSClusterID=$(shell aws rds describe-db-clusters --db-cluster-identifier testdatabase --query 'DBClusters[0].DbClusterResourceId' --output text)" \
		--capabilities CAPABILITY_IAM

.PHONY: deploy
deploy:
	build
	deploy_rds
	init_db
	deploy_api

.PHONY: teardown
teardown:
	aws cloudformation delete-stack --stack-name $(STACK_NAME)
	aws cloudformation delete-stack --stack-name $(STACK_NAME)-api
	clean