import * as codedeploy from '@aws-cdk/aws-codedeploy';
import * as dynamodb from '@aws-cdk/aws-dynamodb';
import * as iam from '@aws-cdk/aws-iam';
import * as lambda from '@aws-cdk/aws-lambda';
import * as ssm from '@aws-cdk/aws-ssm';
import { App, Duration, Stack, StackProps } from '@aws-cdk/core';
      
export class LambdaStack extends Stack {
  public readonly lambdaCode: lambda.CfnParametersCode;
      
  constructor(app: App, id: string, props?: StackProps) {
    super(app, id, props);
      
    this.lambdaCode = lambda.Code.fromCfnParameters();

    // const table = new dynamodb.Table(this, 'Table', {
    //   partitionKey: { 
    //       name: 'name',
    //       type: dynamodb.AttributeType.STRING,
    //   },
    //   readCapacity: 2,
    //   writeCapacity: 2,
    // });

    const executionRole = new iam.Role(this, 'ExecutionRole', {
        assumedBy: new iam.ServicePrincipal('lambda.amazonaws.com'),
        managedPolicies: [
            iam.ManagedPolicy.fromAwsManagedPolicyName("service-role/AWSLambdaBasicExecutionRole"),
            iam.ManagedPolicy.fromAwsManagedPolicyName("AmazonDynamoDBFullAccess"),
        ],
    });
      
    const func = new lambda.Function(this, 'Lambda', {
      code: this.lambdaCode,
      handler: 'main',
      runtime: lambda.Runtime.GO_1_X,
      description: `Function generated on: ${new Date().toISOString()}`,
      role: executionRole,
      timeout: Duration.seconds(120),
      memorySize: 256,
      environment: {
        'DIRECTORY_NAME': 'placeholder',//table.tableName,
        'SLACK_WEBHOOK_URL': ssm.StringParameter.fromStringParameterAttributes(this, 'SlackWebhookUrl', {
            parameterName: 'SLACK_WEBHOOK_URL',
        }).stringValue,
      },
    });
      
    const alias = new lambda.Alias(this, 'LambdaAlias', {
      aliasName: 'dev',
      version: func.currentVersion,
    });
      
    new codedeploy.LambdaDeploymentGroup(this, 'DeploymentGroup', {
      alias,
      deploymentConfig: codedeploy.LambdaDeploymentConfig.ALL_AT_ONCE,
    });
  }
}
