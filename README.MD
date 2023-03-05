# Coding Challenge

Write a script to remove all but the most recent X deployments. The script should take in X as a parameter.

If a deployment is older than X, we will delete the entire folder.

## Questions

1. Where should we run this script?
   This script should be run locally but can run anywhere with access to the Amazon API and the following Environment Variables:

```sh
AWS_ACCESS_KEY_ID
AWS_SECRET_ACCESS_KEY
AWS_DEFAULT_REGION
```

```sh
AWS_REGION
AWS_ENDPOINT
S3_BUCKET
```

**NOTE**: The `AWS_REGION`, `AWS_ENDPOINT`, and `S3_BUCKET` automatically default to presets in the `main.go` program.

2. How should we test the script before running it production?

Please use localstack to setup a test S3 environment. Uncomment lines 168-170 to create test objects.
Set the following as your environment variables to test against the localstack.

```sh
AWS_ACCESS_KEY_ID=test
AWS_SECRET_ACCESS_KEY=test
AWS_DEFAULT_REGION=us-east-1
```

3. If we want to add an additional requirement of deleting deploys older than 30 days while keeping X deployments. What additional changes would you need to make in the script?

I would have to rework the script, so instead of comparing the prefix map based on the deployment count, I would have to compare it to the last modified date before 30 days, only if the number of pairs left is larger than X(deployment count).

## Usage

I built a Go binary package which is essentially the program executable. You can run this without installing Go or any of the dependencies required. This is the fastest way to test the program.

```sh
$./findDeployments -h
Usage of ./findDeployments:
  -delete
        Deletes the S3 objects.
  -deploys int
        Number of most recent deployments to keep.
```

- Run the program and only display the last 30 deployments to delete without deleting them:
  - `./findDeployments -deploys 30`

- Run the program AND delete all deployments that are not the most recent 30 deployments:
  - `./findDeployments -deploys 30 -delete`

## Development Environment

Requirements:

- [Download](https://go.dev/dl/) and install the Go programming language.
  
```sh
go get github.com/aws/aws-sdk-go-v2
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/s3
```

## Assumptions

- We assume the bucket is not version-enabled. If it is then it should mark the objects for deletion by default, which will then delete the current version of the object.
- We assume that all of the data in the buckets are exactly like what was described in the instructions.
