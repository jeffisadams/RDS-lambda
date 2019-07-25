# Supporting codebase for Blog Post
Visit [Here once I publish I will update](https://tenmilesquare.com/) to view the context.

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

## The split
AWS managed services infrastructure is incredibly convenient in a way I had not previously appreciated.  For example, when creating a lambda that pushes to SQS you are only required to add a policy... That's it.  They manage the data on the queue based on your configs.  Similar if you want to write to DynamoDB.  Just a policy.  There are drawbacks.  It becomes a bit opaque where your data actually lives in a DynamoDB table.  But you can connect to it from a lambda by adding a policy.  So when I decided to create a lambda backed by a SQL database instead of DynamoDB I expected a similar level of convenience and instead entered the plumbing center of pain(VPC).

## RDS
The advertised vision of RDS is a Relational database that is \"[easy to set up, operate, and scale](https://aws.amazon.com/rds/faqs/).\"  It acheives this goal in many ways, but there is a very important distinction.  **RDS runs on EC2**.  You pay for compute, and importantly, you must set up the underlying network.  Most tutorials just attach your database to your default VPC.  This is not a bad thing, but it becomes very difficult to add Security groups through [cloudformation](https://aws.amazon.com/cloudformation/) or [SAM](https://aws.amazon.com/serverless/sam/) without a referenceable VPC in your [infrastructure as code](https://en.wikipedia.org/wiki/Infrastructure_as_code).

So now I'm falling down the rabbit hole of setting up a VPC.  And I want to do it using Cloudformation / SAM.  There are many great primers on what VPC is and how to set it up (See links), but I found few that also  had infrastructure as code, and I know why. It's complicated.  So you either get the Hello World VPC tutorial via browser, or the master class that's just the template and sadness.

My goal for this tutorial is to show as simple as possible end to end RDS plus Lambda Plus VPC.  My focus is on the Lambda and the minutia of the Lambda to RDS network connection.  If you are going to run something like this in production, you will need further and deeper instruction on VPC.  I'll include a few tutorials that I liked at the bottom.  I also connect to lambda using a username and password.  AWS recommends using Roles and an Auth Token to connect instead.  I started down this path as the Git history will reveal, but found the scope creaping and my mental health collapsing.  I'll add some auth token tutorials as well at the end.

# The Code

## The Serverless Application Model (SAM) Template
I always recommend utilizing some kind of infrastructure as code.  This stuff is complex and you need to iterate on it.  I promise you if you setup a VPC using the browser, you will forget all the steps next time you need one.  Having a reference and something to step through is critical.

### The VPC
This template is long and has a lot of moving parts.  They are almost all dedicated to the VPC.  My goal is to get to the lambda connection, so I have glossed over an ocean of detail here.  It is also important to keep in mind that AWS did not invent these concepts.  They merely changed the names in confusing ways that made the marketing department happy.  With that in mind, this will be a lot easier to follow if you have some other networking knowledge.  Ensure you know the layered model, and ideally how IP and CIDR route traffic.
Here are the AWS Resources we are building.  There are more VPC related things to build for other use cases, but I am purposely ignoring them to keep us focused on this one use case.
- VPC
  - The abstract grouping of everything else.  Logically isolated network from the rest of AWS.
- Subnets
  - These are both public subnets for simplicity.
  - We are building two because we need two to go in the RDS group which is what tells the RDS what VPC to use in cloudformation.
- Internet Gateway
  - Used to give internet access to resources in your VPC.  Primarily to allow traffic out, but no traffic is allowed in.
- Route Tables
  - IP based route tables.  You can have more than one if you need different rules for each subnets.  Notice the route rule that sends traffic to the internet gateway.  Basically if an IP address shows up and we don't know where it connects.  Head to the open internet to find out.

```
VPC:
  Type: AWS::EC2::VPC
  Properties:
    CidrBlock: 10.1.0.0/16
    EnableDnsSupport: true
    EnableDnsHostnames: true
    Tags:
    - Key: name
      Value:  !Join ['', [!Ref DatabaseName, "-VPC" ]]   
InternetGateway:
  Type: AWS::EC2::InternetGateway
  DependsOn: VPC
AttachGateway:
  Type: AWS::EC2::VPCGatewayAttachment
  Properties:
    VpcId: !Ref VPC
    InternetGatewayId: !Ref InternetGateway

PublicSubnetA:
  Type: AWS::EC2::Subnet
  Properties:
    VpcId: !Ref VPC
    CidrBlock: 10.1.10.0/24
    AvailabilityZone: !Select [ 0, !GetAZs ]
    Tags:
    - Key: Names
      Value: !Sub ${DatabaseName}-PublicA
PublicSubnetB:
  Type: AWS::EC2::Subnet
  Properties:
    VpcId: !Ref VPC
    CidrBlock: 10.1.20.0/24
    AvailabilityZone: !Select [ 1, !GetAZs ]
    Tags:
    - Key: Names
      Value: !Sub ${DatabaseName}-PublicA

PublicRouteTable:
  Type: AWS::EC2::RouteTable
  Properties:
    VpcId: !Ref VPC
PublicRoute:
  Type: AWS::EC2::Route
  DependsOn: AttachGateway
  Properties:
    RouteTableId: !Ref PublicRouteTable
    DestinationCidrBlock: 0.0.0.0/0
    GatewayId: !Ref InternetGateway

PublicSubnetARouteTableAssociation:
  Type: AWS::EC2::SubnetRouteTableAssociation
  Properties:
    SubnetId: !Ref PublicSubnetA
    RouteTableId: !Ref PublicRouteTable

PublicSubnetBRouteTableAssociation:
  Type: AWS::EC2::SubnetRouteTableAssociation
  Properties:
    SubnetId: !Ref PublicSubnetB
    RouteTableId: !Ref PublicRouteTable
```

The goal here is the absolute minumum that will let us build the lambda and RDS in a VPC that can access each other.  There are a lot of other components available for other use cases.  And it's notable that this VPC is all public subnets so this is not suitable for a high security environment.

### Security Group Plumbing to connect to Lambda
The security groups add a second layer of networking rules we can overlay determining which machines / functions can communicate.  We create the Database Security group to allow all outgoing traffic, and incoming traffic from your home IP address along with incoming traffic from the Lambda Security Group.  This is the step that will allow the Lambda process to access the RDS Database without knowing the IP address of the lambda container.

```
LambdaSecurityGroup:
  Type: AWS::EC2::SecurityGroup
  Properties:
    GroupDescription: Allow http to client host
    VpcId: !Ref VPC
    SecurityGroupEgress:
    - IpProtocol: '-1'
      FromPort: -1
      ToPort: -1
      CidrIp: 0.0.0.0/0

DatabaseSecurityGroup:
  Type: AWS::EC2::SecurityGroup
  DependsOn:
    - PublicSubnetA
    - PublicSubnetB
  Properties:
    GroupDescription: Allow http to client host
    VpcId: !Ref VPC
    SecurityGroupIngress:
    # Allows you to log into mysql from home
    - IpProtocol: tcp
      FromPort: 3306
      ToPort: 3306
      CidrIp: !Sub ${YourExternalIp}/32
    # Rule to allow traffic in from resources that are attached to the Lambda Security group
    # This is the specific rule that allows the Lambda to have network access to our RDS cluster
    - IpProtocol: tcp
      FromPort: 3306
      ToPort: 3306
      SourceSecurityGroupId: !GetAtt LambdaSecurityGroup.GroupId
    SecurityGroupEgress:
    - IpProtocol: "-1"
      FromPort: -1
      ToPort: -1
      CidrIp: 0.0.0.0/0
```

### Ok Phew... RDS time
By contrast this is dead simple.  Three resources to add for a boilerplate version.
- The abstract cluster
- One instance
- The subnet group
We choose the VPC using the subnet group, otherwise it just attaches to the default VPC.  The default VPC is suitable for lots of things, but makes Template created security groups difficult.  Right here is where the zero to cluster and api breaks down without the VPC.  The instance will map one to one with an EC2 instance and apply the network rules we spent so much time with above via two config settings.  Obviously it is possible to add more instances, and we would need redundancy for production, but here's a start.

One note:  Use SSM parameters to store your usernames / passwords if and when you do this in production.  I again wanted to keep things as simple as possible, but this is another place where you need higher security in the actual implementation for production. (Tutorial for SSM)[https://aws.amazon.com/blogs/mt/integrating-aws-cloudformation-with-aws-systems-manager-parameter-store/]

```
RDSDatabase:
  Type: AWS::RDS::DBCluster
  Properties:
    DBClusterIdentifier: !Ref DatabaseName
    Engine: aurora
    EnableIAMDatabaseAuthentication: true
    MasterUsername: !Ref AdminUserName
    MasterUserPassword: !Ref AdminUserPassword
    DBSubnetGroupName: !Ref PublicSubnetGroup
    VpcSecurityGroupIds:
      - !GetAtt DatabaseSecurityGroup.GroupId
RDSInstance1:
  Type: AWS::RDS::DBInstance
  DependsOn:
    - PublicSubnetA
  Properties:
    DBClusterIdentifier: !Ref RDSDatabase
    # Customize this to the size machine you want
    DBInstanceClass: db.r4.large
    Engine: aurora
    PubliclyAccessible: true
    DBSubnetGroupName: !Ref PublicSubnetGroup
```

Know that your RDS Cluster is routable via the internet.  With the correct Security group setup, you can directly access this database.  That's helpful for debugging and data cleanup, but can be dicey depending on the data inside and your security profile.  It is possible to create private subnets and make lambda able to access the DB cluster and not have the DB cluster routable from the outside.

### API Gateway
This is a simple lambda in an API Gateway that just retrieves the data from the DB.  We pass in the config using the environment.  Generally we would want to use token Auth instead, but this post is already too long so I link below to tutorials on how to add that.  

```
ServiceApi:
  Type: AWS::Serverless::Api
  Properties:
    Name: ServiceApi
    StageName: !Ref Version
LambdaFunction:
  Type: AWS::Serverless::Function
  Description: Function to get the data out of the database
  Properties:
    Handler: ./dist/main
    Timeout: 30
    Policies:
      - AWSLambdaVPCAccessExecutionRole
    Environment:
      Variables:
        DB_SERVICE_USER: !Ref ServiceUserName
        DB_SERVICE_PASSWORD: !Ref ServiceUserPassword
        DB_HOST: !GetAtt RDSDatabase.ReadEndpoint.Address
        DB_NAME: !Ref DatabaseName
        DB_TABLE_NAME: !Ref DatabaseTableName
    VpcConfig:
      SecurityGroupIds:
        - !GetAtt LambdaSecurityGroup.GroupId
      SubnetIds:
        - !Ref PublicSubnetA
        - !Ref PublicSubnetB
    Events:
      CreateUser:
        Type: Api
        Properties:
          Path: /transactions
          RestApiId: !Ref ServiceApi
          Method: Get
```

The critical piece is that the lambda runs in the same VPC as the DB allowing us to access it via Security Group Rules.  Since our Database is technically public as well, we could setup access from the outside, but then we would not be able to protect the database using security groups.  And the Lambda by definition does not have a consistent IP address so there's no IP based rule we can apply.  Instead we add it to the security group in the VPC.


# The Runtime Guts (But first some init)

We now have all the AWS resources setup.  However the API endpoint will fail because it relies on a user that does not yet exist.

```
Environment:
  Variables:
    DB_SERVICE_USER: !Ref ServiceUserName
    DB_SERVICE_PASSWORD: !Ref ServiceUserPassword
    DB_HOST: !GetAtt RDSDatabase.ReadEndpoint.Address
    DB_NAME: !Ref DatabaseName
    DB_TABLE_NAME: !Ref DatabaseTableName
```

So I created a script to create a user.  This is the part that typically would have token authentication, but let's focus on the connectivity and I'll let better tutorials show the token auth steps (See Below).

### Init and user creation
We need a user, a database, a table, and some data.  So let's wave our hands and run the script `make init_db`.  This runs a script in Go with command line Env variables.  My original intentionwas to make a custom resource to do this, but found the Go library for custom resources infuriating.  If you have any error of any kind, Cloudformation will wait 60 minutes to see if things stabilize.  Also, since the Lambda runs in a VPC.  Tearing down this Cloudformation template also takes about 30 minutes.  I have lost days of my life to this issue.  Secondly, that Lambda would have environment variables with your master username and password which didn't seem like a good idea.

```
db.MustExec(fmt.Sprintf(`CREATE DATABASE IF NOT EXISTS %s`, dbName))
db.MustExec(fmt.Sprintf(`USE %s`, dbName))

db.MustExec(fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s';", dbServiceUser, dbServicePassword))
db.MustExec(fmt.Sprintf("GRANT ALL ON %s.* TO '%s'@'%%';", dbName, dbServiceUser))
// db.MustExec("FLUSH PRIVILEGES;")

db.MustExec(fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s(
  id INT NOT NULL AUTO_INCREMENT,
  email varchar(255) DEFAULT NULL,
  date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  CONSTRAINT UNIQUE(email)
);`, dbTableName))

db.MustExec(fmt.Sprintf(`INSERT INTO %s (email) VALUES ("bob@example.com");`, dbTableName))
db.MustExec(fmt.Sprintf(`INSERT INTO %s (email) VALUES ("jane@example.com");`, dbTableName))
```

### The actual lambda
This is pretty straightforward, but relies heavily on all that came before.

```
func handleCrudRequest() (events.APIGatewayProxyResponse, error) {
	dbTableName := os.Getenv("DB_TABLE_NAME")
	transactions := []UserTransaction{}
	err := db.Select(&transactions, fmt.Sprintf(`SELECT * from %s`, dbTableName))
	if err != nil {
		fmt.Println(err)
	}

	out, serializationErr := json.Marshal(transactions)
	if serializationErr != nil {
		fmt.Println(serializationErr)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(out),
	}, nil
}

func main() {
	lambda.Start(handleCrudRequest)
}

func init() {
	// Login to the DB
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_SERVICE_USER")
	dbPassword := os.Getenv("DB_SERVICE_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	connectionStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=10s", dbUser, dbPassword, dbHost, "3306", dbName)
	var err error
	db, err = sqlx.Connect("mysql", connectionStr)
	if err != nil {
		log.Fatalln(err)
	}
}

```

We use environment variables to pass the table and user settings for the db connection.  Then pull from our table and return it.  Before this will successfully run, ensure you ran `make init_db`

# How to test this
In Cloudformation, look for the output tab in your template and follow that link (Structured similarly to below but with your ids).  You should see a json list of two objects returned:

RUN `curl https://xxxxxxxxxx.execute-api.us-xxxxxxxx.amazonaws.com/v1/transactions`
```
[
  {
    id: 1,
    email: "bob@example.com",
    date: "2019-07-12 19:09:25"
  },
  {
    id: 2,
    email: "jane@example.com",
    date: "2019-07-12 19:09:25"
  }
]

```

# Summary of what we built
- A functioning VPC with two public Subnets in two Availability Zones.
- The plumbing in our VPC to give access to the RDS database
  - Security Groups that allow traffic without using an IP address
- An RDS Cluster with one instance running in one Availability zone (we made two zones since you have to have two zones to create a subnet group for RDS).
- A DB init script that runs locally to create:
  - A database
  - A user we can use to log in with
  - A table
  - Data that we can pull
- An API Gateway Lambda function that runs a simple query to the Database running in the same VPC and outputs it

## Don't forget to tear down the infrastructure on this one
Since this template runs an EC2 machine, the price adds up a lot quicker than DynamoDB or Queues or other fully managed services.  Run `make teardown` to clean it all up.

## Reading List
- General Links
  - [Virtual Private Cloud (VPC)](https://aws.amazon.com/vpc/)
- VPC Tutorials
  - [VPC Routing](https://medium.com/@mda590/aws-routing-101-67879d23014d)
- AWS RDS Token Auth Connection Tutorials
  - [Token Authentication](https://aws.amazon.com/premiumsupport/knowledge-center/users-connect-rds-iam/)
  - [AWS Token Auth Example](https://github.com/aws/aws-sdk-go/blob/master/example/service/rds/rdsutils/authentication/iam_authentication.go)
  - [The one that had the working RDSUtils code in GO](https://luktom.net/en/e1544-aws-lambda-and-mysql-iam-authentication-in-go)
  - [An ongoing question](https://stackoverflow.com/questions/48138267/unable-to-access-rds-database-via-iam-authentication)
