#!/bin/sh

PARAMETER_VALUE=$(aws ssm get-parameter --name ProxyEnv --query Parameter.Value --output text --region ap-northeast-1)
# .envファイルに書き込む
echo "PARAMETER_NAME=$PARAMETER_VALUE" >> /home/ec2-user/app/.env
docker image build -t proxy:latest /home/ec2-user/app
docker run --env-file /home/ec2-user/app/.env -p 80:80 proxy:latest > /dev/null 2>&1 &
