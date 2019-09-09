# Supporting codebase for Blog Post
Visit [Here](https://tenmilesquare.com/rds-and-stepping-into-the-plumbing-center-of-pain/) to view the context.

## To Run
- Set the following environment variables
  - STACK_NAME
  - STACK_BUCKET (this bucket must exist in your AWS environment)
  - YOUR_IP What is your public IP so it gets added to the security group
    - This is the only way you'll be able to access your database
- run the following in succession:
  - `make build`
  - `make deploy_rds`
  - `make init_db`
- run `curl {{Url from the output tab of your cloudformation template}}`
  - You should see two records retrieved and serialized from the database
    - bob@example.com
    - jane@example.com

PRs and suggestions are welcome.
