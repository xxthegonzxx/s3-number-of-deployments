package actions

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// BucketBasics encapsulates the Amazon Simple Storage Service (Amazon S3) actions.
// It contains S3Client, an Amazon S3 service client that is used to perform bucket
// and object actions.
type BucketActions struct {
	S3Client *s3.Client
}

// Create Bucket
func (actions BucketActions) CreateBucket(bucketName string) {
	actions.S3Client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
}

// ListBuckets returns the bucket name and the date of creation.
func (actions BucketActions) ListBuckets() {
	result, err := actions.S3Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if err != nil {
		log.Fatalf("No buckets found. Error: %v", err)
	}
	fmt.Println("Buckets:")
	for _, bucket := range result.Buckets {
		fmt.Println(*bucket.Name + ": " + bucket.CreationDate.Format("2006-01-02 15:04:05 Monday"))
	}
}

// UploadFile reads from a file and puts the data into an object in a bucket.
func (actions BucketActions) UploadFile(bucketName string, objectKey string, fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Couldn't open file %v to upload. Here's why: %v\n", fileName, err)
	} else {
		defer file.Close()
		_, err := actions.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectKey),
			Body:   file,
		})
		if err != nil {
			log.Printf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
				fileName, bucketName, objectKey, err)
		}
	}
	return err
}

func (actions BucketActions) CreateObjects(bucketName string, objectKey string) {
	body := []byte("Hello World")
	actions.S3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:             aws.String(bucketName),
		Key:                aws.String(objectKey),
		Body:               bytes.NewReader(body),
		ContentType:        aws.String("application/text"),
		ContentDisposition: aws.String("attachment"),
	})

}

// ListObjects lists the objects in a bucket.
func (actions BucketActions) ListObjects(bucketName string) ([]types.Object, error) {
	result, err := actions.S3Client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})
	var contents []types.Object
	if err != nil {
		log.Printf("Couldn't list objects in bucket %v. Here's why: %v\n", bucketName, err)
	} else {
		contents = result.Contents
	}
	return contents, err
}

// DeleteObjects deletes a list of objects from a bucket.
func (actions BucketActions) DeleteObjects(bucketName string, objectKeys []string) error {
	var objectIds []types.ObjectIdentifier
	for _, key := range objectKeys {
		objectIds = append(objectIds, types.ObjectIdentifier{Key: aws.String(key)})
	}
	_, err := actions.S3Client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{Objects: objectIds},
	})
	if err != nil {
		log.Printf("Couldn't delete objects from bucket %v. Here's why: %v\n", bucketName, err)
	}
	return err
}
