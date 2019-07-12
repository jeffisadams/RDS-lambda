# Supporting codebase for Blog Post
Visit (Here once I publish I will update)[https://tenmilesquare.com/] to view the context.

## To Run
- Set the following environment variables
  - STACK_NAME
  - STACK_BUCKET (this bucket must exist in your AWS environment)
  - YOUR_IP What is your public IP so it gets added to the security group
    - This is the only way you'll be able to access your database
- run `make deploy`
  - This will setup the database and run a custom function to create the database
- run `make deploy_api`
  - This requires that the previous stack has completed initializing
- run `curl {{Url to your api}}/transactions`
  - No Authentication at this time
  - This will return two data transaction records for bob and jane at example.com

PRs and suggestions are welcome.


--- Start the Blog Content

# RDS and stepping into the plumbing center of pain:
## Lambda example in a VPC backed by RDS

Special Thanks to Jason Mao for the help reviewing the network VPC setup.

## TLDR;
  Code Located [here](https://github.com/jeffisadams/RDS-lambda).

## Pre-requisites
- [AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)
- [AWS Serverless Application Model CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
- [Go](https://golang.org/doc/install)
- [Command line access to your environment](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html)

### To Run
- Set the following environment variables
  - STACK_NAME
  - STACK_BUCKET (this bucket must exist in your AWS environment)
  - YOUR_IP What is your public IP so it gets added to the security group
    - This is the only way you'll be able to access your database
- run `make deploy`
  - This will setup the database and run a custom function to create the database
- run `make deploy_api`
  - This requires that the previous stack has completed initializing
- run `curl {{Url to your api}}/transactions`
  - No Authentication at this time
  - This will return two data transaction records for bob and jane at example.com

## The split
AWS managed services infrastructure is incredibly convenient in a way I had not previously appreciated.  Say you want a lambda that pushes to a a SQS queue of some kind.  You add a policy... That's it.  They manage the data on the queue based on your configs.  Similar if you want to write to DynamoDB.  Just a policy.  There are drawbacks.  It becomes a bit opaque where your data actually lives in a DynamoDB table.  It's usually in a region, it's replicated across availability zones.  It's non relational.  But you can connect to it from a lambda by adding a policy to your lambda.  So when I decided to create a lambda backed by a SQL database instead of DynamoDB I expected a similar level of convenience and instead entered the plumbing center of pain(VPC).

## RDS
The advertised vision of RDS is a Relational database that is \"[easy to set up, operate, and scale](https://aws.amazon.com/rds/faqs/).\"  It acheives this goal in many ways, but there is a very important distinction.  RDS runs on EC2.  You pay for compute, and importantly, you must set up the underlying network.  Most tutorials just attach your database to your default VPC.  This is not a bad thing, but it becomes very difficult to add Security groups through [cloudformation](https://aws.amazon.com/cloudformation/) or [SAM](https://aws.amazon.com/serverless/sam/) without a referenceable VPC in your [infrastructure as code](https://en.wikipedia.org/wiki/Infrastructure_as_code).

So now I'm falling down the rabbit hole of setting up a VPC.  And I want to do it using Cloudformation / SAM.  There are many great primers on what VPC is and how to set it up (See links), but I found few that also  had infrastructure as code, and I know why. It's complicated.  So you either get the Hello World VPC tutorial via browser, or the master class that's just the template and sadness.

My goal for this tutorial is to show as simple as possible end to end RDS plus Lambda Plus VPC.  My focus is on the Lambda and the minutia of the Lambda to RDS connection.  If you are going to run something like this in production, you will need further and deeper instruction on VPC.  I'll include a few tutorials that I liked at the bottom.

# The Code

## The Template
I always recommend utilizing some kind of infrastructure as code.  This stuff is complex and you need to iterate on it.  I promise you if you setup a VPC using the browser, you will forget all the steps next time you need one.  Having a reference and something to step through is critical.

### The VPC
This template is long and has a lot of moving parts.  They are almost all dedicated to the VPC.  My goal is to get to the lambda connection, so I have glossed over an ocean of detail here.  It is also important to keep in mind that AWS did not invent these concepts.  They merely changed the names in confusing ways that made the marketing department happy.  With that in mind, this will be a lot easier to follow if you have done some other networking knowledge.  Ensure you know the layered model, and ideally how IP and CIDR route traffic.
Here are the AWS Resources we are building.  There are more VPC related things to build for other use cases, but I am purposely ignoring them to keep us focused on this one use case.
- VPC
  - The abstract grouping of everything else.  Logically isolated network from the rest of AWS.
- Subnets
  - Private and Public subnets. What's the difference you ask.  Good question, [waves hands] (https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Scenario2.html).
  - We are building two because we need two to go in the RDS group which is what tells the RDS what VPC to use in cloudformation.
- Internet Gateways
  - Used to give internet access to resources in your VPC.  Primarily to allow traffic out, but no traffic is allowed in.
  - Importantly there is also a Nat Gateway.  This is used if you need customized routing into your VPC (EX: if you need an EC2 machine that is addressable via IP from the outside).  In our case, RDS will handle this for us making the canonical DB endpoint and mapping that to our EC2 instance
- Route Tables
  - IP based route tables.  You can have more than one if you need different rules for each subnets.  Notice the route rule that sends traffic to the internet gateway.  Basically if an IP address shows up and we don't know where it connects.  Head to the open internet to find out

'''

'''

### Ok Phew... RDS time
By contrast this is dead simple.  Three resources to add for a boilerplate version.  The abstract cluster, one instance, and the subnet group so we choose the VPC (otherwise it just attaches to the default).  The instance will map one to one with an EC2 instance and apply the network rules we spent so much time with above via two config settings.  Obviously it is possible to add more instances, and we would need redundancy for production, but here's a start.

```

```

Know that your RDS is on the internet.  With the correct Security group setup, you can directly access this database.  That's helpful for debugging and data cleanup, but can be dicey depending on the data inside and your security profile.

### Security Group Plumbing to connect to Lambda
The security groups add a second layer of rules we can overlay determining which boxes can communicate.  We create the Databse Security group to allow all outgoing traffic, and incoming traffic from your home IP address along with incoming traffic from the Lambda Security Group.  This is the step that will allow the Lambda process to access the RDS Database without knowing the IP address of the lambda container.

```

```

### Deploy Lambda Custom Resource
This lambda is a custom SAM resource.  The whole goal is bootstrapping.  We use the database credentials from our environment to run once and create a database and a role based user.  We build to items: The function that will run, and a custom SAM recource that calls the deploy Lambda on the Cloudformation deploy event.  IMPORTANT!!! The custom function has specific rules about the Lambda responding with a success event and not failing to run.  If the lambda fails to run, Your Cloudformation deploy has a 60 minute timeout!  This means an hour of powerlessness to do ANYTHING with regards to this running process.  Learn from my mistakes and make sure to test as many steps as you can before spending a day angrily staring at the cloudformation status screen.

```

```

### Deploy the Lambda for the API
Now that we have a database that has been initialized, we can connect to it using our created user and pull surface our database info to the API Gateway.  Take note of the guts to connect using a role based AWS session connection instead of a username and password.

```

```

# Summary of what we built
- A functioning VPC with two public Subnets in two Availability Zones.
- The plumbing in our VPC to give access to the RDS database
- An RDS Cluster with one instance running in one Availability zone (we made two zones since you have to have two zones to create a subnet group for RDS).
- A Custom SAM resource that runs an initialization function to create a database, a table and a user.
- A Lambda function that runs a simple query to the Database running in the same VPC and outputs it to our API Gateway

## Don't forget to tear down the infrastructure on this one
Since this template runs an EC2 machine, the price adds up a lot quicker than DynamoDB or Queues or other fully managed services.

Run `make teardown` to clean it all up.

## This is not ready for production
Most notably, you should not use the master user or master password to login from your lambda.  My initial goal was to create a Deploy lambda that creates the ability to log in via a role to the instance instead, but I don't know that I'll have time to include that.  If not, I'll do a follow up on custom resources.

## Reading List
- VPC Links
  - [Virtual Private Cloud (VPC)](https://aws.amazon.com/vpc/)
  - [VPC Routing](https://medium.com/@mda590/aws-routing-101-67879d23014d)
- AWS RDS Connection Tutorials
  - [Token Authentication](https://aws.amazon.com/premiumsupport/knowledge-center/users-connect-rds-iam/)
  - [AWS Token Auth Example](https://github.com/aws/aws-sdk-go/blob/master/example/service/rds/rdsutils/authentication/iam_authentication.go)
  - [The one that had the working RDSUtils code in GO](https://luktom.net/en/e1544-aws-lambda-and-mysql-iam-authentication-in-go)
- Custom Resource Links