import * as codedeploy from '@aws-cdk/aws-codedeploy';
import * as lambda from '@aws-cdk/aws-lambda';
import * as ssm from '@aws-cdk/aws-ssm';
import { App, Stack, StackProps } from '@aws-cdk/core';
      
export class LambdaStack extends Stack {
  public readonly lambdaCode: lambda.CfnParametersCode;
      
  constructor(app: App, id: string, props?: StackProps) {
    super(app, id, props);
      
    this.lambdaCode = lambda.Code.fromCfnParameters();
      
    const func = new lambda.Function(this, 'Lambda', {
      code: this.lambdaCode,
      handler: 'main',
      runtime: lambda.Runtime.GO_1_X,
      description: `Function generated on: ${new Date().toISOString()}`,
      environment: {
        'DIRECTORY_NAME': ssm.StringParameter.fromStringParameterAttributes(this, 'DirectoryName', {
            parameterName: 'DIRECTORY_NAME',
        }).stringValue,
        'SLACK_WEBHOOK_URL': ssm.StringParameter.fromSecureStringParameterAttributes(this, 'SlackWebhookUrl', {
            parameterName: 'SLACK_WEBHOOK_URL',
            version: 1,
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
