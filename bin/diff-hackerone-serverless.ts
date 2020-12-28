#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { DiffHackeroneServerlessStack } from '../lib/diff-hackerone-serverless-stack';

const app = new cdk.App();
new DiffHackeroneServerlessStack(app, 'DiffHackeroneServerlessStack');
