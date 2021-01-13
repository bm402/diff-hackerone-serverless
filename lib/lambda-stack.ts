import * as codedeploy from '@aws-cdk/aws-codedeploy';
import * as dynamodb from '@aws-cdk/aws-dynamodb';
import * as event from '@aws-cdk/aws-events';
import * as iam from '@aws-cdk/aws-iam';
import * as lambda from '@aws-cdk/aws-lambda';
import * as ssm from '@aws-cdk/aws-ssm';
import * as target from '@aws-cdk/aws-events-targets';
import { App, Duration, Stack, StackProps } from '@aws-cdk/core';

/**
 * Defines a CloudFormation stack to create and maintain the AWS serverless infrastructure required to
 * run the application.
 */
export class LambdaStack extends Stack {
  public readonly lambdaCode: lambda.CfnParametersCode;
      
  constructor(app: App, id: string, props?: StackProps) {
    super(app, id, props);
    
    // Retrieves Lambda application from the CfnParameters (defined in the Deploy stage of the pipeline stack)
    this.lambdaCode = lambda.Code.fromCfnParameters();

    // DynamoDB table for programs and assets
    const table = new dynamodb.Table(this, 'Table', {
      partitionKey: { 
          name: 'name',
          type: dynamodb.AttributeType.STRING,
      },
      readCapacity: 2,
      writeCapacity: 2,
    });

    // Lambda execution role with DynamoDB access
    const executionRole = new iam.Role(this, 'ExecutionRole', {
        assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
        managedPolicies: [
            iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole"),
            iam.ManagedPolicy.fromAwsManagedPolicyName("AmazonDynamoDBFullAccess"),
        ],
    });
    
    // Main Lambda function
    const func = new lambda.Function(this, 'Lambda', {
      code: this.lambdaCode,
      handler: 'main',
      runtime: lambda.Runtime.GO_1_X,
      description: `Function generated on: ${new Date().toISOString()}`,
      role: executionRole,
      timeout: Duration.seconds(120),
      memorySize: 256,
      environment: {
        'DIRECTORY_NAME': table.tableName,
        'SLACK_WEBHOOK_URL': ssm.StringParameter.fromStringParameterAttributes(this, 'SlackWebhookUrl', {
            parameterName: 'SLACK_WEBHOOK_URL',
        }).stringValue,
      },
    });
    
    // Lambda alias
    const alias = new lambda.Alias(this, 'LambdaAlias', {
      aliasName: 'dev',
      version: func.currentVersion,
    });

    // EventBridge rule for running the Lambda function every 15 minutes
    const lambdaTarget = new target.LambdaFunction(func)
    new event.Rule(this, 'ScheduleRule', {
        schedule: event.Schedule.rate(Duration.minutes(15)),
        targets: [
            lambdaTarget,
        ]
    });
    
    // Deployment group for pipeline
    new codedeploy.LambdaDeploymentGroup(this, 'DeploymentGroup', {
      alias,
      deploymentConfig: codedeploy.LambdaDeploymentConfig.ALL_AT_ONCE,
    });
  }
}
