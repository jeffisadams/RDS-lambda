# Supporting codebase for Blog Post
Visit (Here once I publish I will update)[https://tenmilesquare.com/] to view the context.

## To Run
- Set the following environment variables
  - STACK_NAME
  - STACK_BUCKET (this bucket must exist in your AWS environment)
- run `make deploy`
- run `curl {{Url to your api}}/`
  - No Authentication at this time

PRs and suggestions are welcome.




--- Start the Blog Content

# RDS and stepping into the plumbing center of pain:
## Lambda example in a VPC backed by RDS

Special Thanks to Jason Mao for the help reviewing the network VPC setup.

This is a bit of rant, but I promise there is an end to end guide on setting up a Lambda function that writes to RDS Aurora SQL embedded.


## Terms
- [VPC](https://aws.amazon.com/vpc/) "Virtual Private Cloud"
  - Managed and provisioned network space where you get to set the rules.
- [Subnet]

## The split
AWS managed services infrastructure is incredibly convenient in a way I had not previously appreciated.  Say you want a lambda that pushes to a a SQS queue of some kind.  You add a policy... That's it.  They manage the data on the queue based on your configs.  Similar if you want to write to DynamoDB.  Just a policy.  There are drawbacks.  It becomes a bit opaque where your data actually lives in a DynamoDB table.  It's usually in a region, it's replicated across availability zones.  It's non relational.  But you can connect to it from a lambda by adding a policy to your lambda.  So when I decided to create a lambda backed by a SQL database instead of DynamoDB I expected a similar level of convenience and instead entered the plumbing center of pain(VPC).

## RDS
The advertised vision of RDS is a Relational database that is \"[easy to set up, operate, and scale](https://aws.amazon.com/rds/faqs/).\"  It acheives this goal in many ways, but there is a very important distinction.  RDS runs on EC2.  You pay for compute, and importantly, you must set up the underlying network.  Most tutorials just attach your database to your default VPC.  This is not a bad thing, but it becomes very difficult to add Security groups through [cloudformation](https://aws.amazon.com/cloudformation/) or [SAM](https://aws.amazon.com/serverless/sam/) without a referenceable VPC in your [infrastructure as code](https://en.wikipedia.org/wiki/Infrastructure_as_code).

So now I'm falling down the rabbit hole (Rat Hole) of setting up a VPC.  And I want to do it using Cloudformation / SAM.  There are many great primers on what VPC is and how to set it up (See links), but I found few that also  had infrastructure as code, and I know why. It's complicated.  So you either get the Hello World VPC tutorial via browser, or the master class that's just the template and sadness.

The goal is to instead show as simple as possible end to end RDS plus Lambda Plus VPC.  My focus is on the Lambda and the minutia of the Lambda to RDS connection.  If you are going to run something like this in production, you will need further and deeper instruction on VPC.  I'll include a few tutorials that I liked at the bottom.

## TLDR;
  Code Located [here](https://github.com/jeffisadams/).

  Steps:
    - Set Env var STACK_BUCKET to an existing bucket in your infrastructure


# The Code


## The Template

### The VPC
This template is long and has a lot of moving parts.  They are almost all dedicated to the VPC.  Here is a list of what is needed for the VPC.  My goal is to get to the lambda connection, so I have glossed over an ocean of detail here.  This is where Jason's help has been invaluable:
- VPC
  - The abstract 
- Subnets
  - Private and Public subnets. What's the difference you ask.  Good question, [waves hands] (https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Scenario2.html).
- Internet Gateways
  - Primarily to allow traffic out, but not in (Private Subnet).  Or make a dedicated corner of the internet accessible behind a NAT(Public Subnet).
- Route Tables
  - 




### Ok Phew... RDS time
By contrast this is dead simple.  Two items to add for a boilerplate version.  The abstract cluster, and one instance.  The instance will map one to one with and EC2 instance and apply the network rules we spent so much time with above via two config settings.

```

```



# Some notes at the end

## Don't forget to tear down the infrastructure on this one
Since this template runs an EC2 machine, the price adds up a lot quicker than DynamoDB or Queues or other fully managed services.

Run `make teardown` to clean it all up.

## This is not ready for production
Most notably, you should not use the master user or master password to login from your lambda.  My initial goal was to create a Deploy lambda that creates the ability to log in via a role to the instance instead, but I don't know that I'll have time to include that.

## Reading List
- General Links
  - [Virtual Private Cloud (VPC)](https://aws.amazon.com/vpc/) 
- VPC Tutorials
