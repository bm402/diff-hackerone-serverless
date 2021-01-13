# Diff-HackerOne serverless
A serverless application that sends Slack notifications to the user when HackerOne updates their bug bounty program directory. This helps you be one of the first people to analyse a new program or an updated asset of an existing program.

The application uses the AWS CDK to create all of the serverless infrastructure required on AWS, as well as a CI/CD pipeline which updates the deployment when when an update is made to this repository.

## Prerequisites
To run the deployment, the following software is required:
 * Node.js
 * AWS CLI
 * AWS CDK

## Usage

### Configure AWS credentials
Use the AWS CLI to configure your AWS credentials:
```
$ aws configure
```

### Bootstrap CDK
Provision AWS resources that will be used by the CDK:
```
$ cdk bootstrap
```

### Install dependencies
Install the dependencies used by the CDK and the application:
```
$ npm install
```

### Add GitHub access token to AWS Secrets Manager
To allow the pipeline access to your GitHub repository, add a GitHub access token with full `repo` and `repo_hook` permissions to the AWS Secrets Manager. Name the secret `GITHUB_ACCESS_TOKEN`.

### Add Slack webhook URL to AWS SSM Parameter Store
In the AWS SSM Parameter Store, place your Slack webhook URL in a parameter named `SLACK_WEBHOOK_URL`. Ensure this parameter is a plain string and not a secure string; secure strings are not currently supported by the CDK.

### Run the deployment
Use the CDK to provision the required AWS infrastructure and build the application:
```
$ cdk deploy PipelineDeployingLambdaStack
```

### Other useful commands

 * `npm run build`   compile typescript to js
 * `npm run watch`   watch for changes and compile
 * `npm run test`    perform the jest unit tests
 * `cdk diff`        compare deployed stack with current state
 * `cdk synth`       emits the synthesized CloudFormation template

## Deployed AWS infrastructure
The following AWS infrastructure will be deployed by the Lambda stack:
 * **Lambda function**: The main application which interacts with HackerOne through a GraphQL endpoint and sends Slack notifications with details of any changes to the HackerOne program directory
 * **DynamoDB table**: Stores a local copy of the HackerOne directory
 * **Lambda execution role**: The role used by the Lambda function to execute and to access the DynamoDB table
 * **EventBridge rule**: A rule that triggers the execution of the Lambda function every 15 minutes

The following AWS infrastructure will be deployed by the Pipeline stack:
 * **CodeCommit pipeline stage**: Retrieves the source code from GitHub
 * **CodeBuild pipeline stage**: Builds both the Lambda function and the Lambda stack from source
 * **CodeDeploy pipeline stage**: Deploys the Lambda stack

## References
[https://docs.aws.amazon.com/cdk/latest/guide/codepipeline_example.html](https://docs.aws.amazon.com/cdk/latest/guide/codepipeline_example.html)
