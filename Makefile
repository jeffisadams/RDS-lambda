STACK_NAME := RDS-lambda-test-stack
DB_NAME := testdatabase
DB_TABLE_NAME := transactions
DB_SERVICE_USER := lambda_user
DB_SERVICE_PASSWORD := insecure_1234
DB_ADMIN_USERNAME := admin
DB_ADMIN_PASSWORD := event_more_insecure_1234

ifeq ($(STACK_BUCKET),)
$(error You must specify STACK_BUCKET)
endif

ifeq ($(YOUR_IP),)
$(error You must specify YOUR_IP)
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
		--parameter-overrides "DatabaseName=$(DB_NAME)" "DatabaseTableName=$(DB_TABLE_NAME)" "ServiceUserName=$(DB_SERVICE_USER)" "ServiceUserPassword=$(DB_SERVICE_PASSWORD)" "AdminUserName=$(DB_ADMIN_USERNAME)" "AdminUserPassword=$(DB_ADMIN_PASSWORD)" "YourExternalIp=$(YOUR_IP)"

.PHONY: teardown
teardown:
	aws cloudformation delete-stack --stack-name $(STACK_NAME)
	clean

.PHONY: init_db
init_db:
	./dist/init $(DB_ADMIN_USERNAME) $(DB_ADMIN_PASSWORD) $(shell aws rds describe-db-clusters --db-cluster-identifier $(DB_NAME) --query 'DBClusters[0].Endpoint' --output text) $(DB_NAME) $(DB_TABLE_NAME) $(DB_SERVICE_USER) $(DB_SERVICE_PASSWORD)
