#!/bin/bash
# Quick AWS credentials verification script

set -e

echo "🔐 Testing AWS Credentials for S3 Event Exporter"
echo "=============================================="

# Check if AWS CLI is available
if ! command -v aws &> /dev/null; then
    echo "⚠️  AWS CLI not found. Installing is recommended but not required."
    echo "You can still deploy if you have AWS credentials set in environment variables."
fi

# Check environment variables
echo "📋 Checking AWS environment variables..."
if [[ -n "$AWS_ACCESS_KEY_ID" ]]; then
    echo "✅ AWS_ACCESS_KEY_ID is set"
else
    echo "❌ AWS_ACCESS_KEY_ID is not set"
fi

if [[ -n "$AWS_SECRET_ACCESS_KEY" ]]; then
    echo "✅ AWS_SECRET_ACCESS_KEY is set"
else
    echo "❌ AWS_SECRET_ACCESS_KEY is not set"
fi

if [[ -n "$AWS_REGION" ]]; then
    echo "✅ AWS_REGION is set to: $AWS_REGION"
else
    echo "⚠️  AWS_REGION not set, will default to us-west-2"
    export AWS_REGION=us-west-2
fi

# Test AWS credentials if AWS CLI is available
if command -v aws &> /dev/null && [[ -n "$AWS_ACCESS_KEY_ID" && -n "$AWS_SECRET_ACCESS_KEY" ]]; then
    echo ""
    echo "🧪 Testing AWS credentials..."
    if aws sts get-caller-identity &> /dev/null; then
        echo "✅ AWS credentials are valid"
        aws sts get-caller-identity
        
        echo ""
        echo "🧪 Testing SQS access..."
        QUEUE_URL="https://sqs.us-west-2.amazonaws.com/305018987196/s3-event-exporter-dev-s3-event-queue"
        if aws sqs get-queue-attributes --queue-url "$QUEUE_URL" --attribute-names All &> /dev/null; then
            echo "✅ SQS queue is accessible"
        else
            echo "❌ SQS queue is not accessible or doesn't exist"
            echo "   Queue: $QUEUE_URL"
            echo "   This might be expected if the queue doesn't exist yet"
        fi
        
        echo ""
        echo "🧪 Testing S3 bucket access..."
        BUCKET_NAME="s3-event-exporter-dev-test-bucket"
        if aws s3 ls "s3://$BUCKET_NAME" &> /dev/null; then
            echo "✅ S3 bucket is accessible"
        else
            echo "❌ S3 bucket is not accessible or doesn't exist"
            echo "   Bucket: $BUCKET_NAME"
            echo "   This might be expected if the bucket doesn't exist yet"
        fi
    else
        echo "❌ AWS credentials are not valid"
        echo "Please check your AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY"
        exit 1
    fi
else
    echo "⚠️  Skipping AWS credential test (AWS CLI not available or credentials not set)"
fi

echo ""
echo "🚀 Ready to deploy! Run: ./scripts/deploy-local.sh"
