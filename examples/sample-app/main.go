package main

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func main() {
	const (
		S3            = "s3"
		DefaultRegion = "default"
	)

	accessKey := os.Getenv("S3_ACCESS_KEY_1")
	secretKey := os.Getenv("S3_SECRET_KEY_1")
	bucketName := os.Getenv(("S3_BUCKET_NAME_1"))
	accessKey2 := os.Getenv("S3_ACCESS_KEY_2")
	secretKey2 := os.Getenv("S3_SECRET_KEY_2")
	bucketNameTwo := os.Getenv(("S3_BUCKET_NAME_2"))
	endpointStr := os.Getenv("S3_ENDPOINT")
	regionStr := os.Getenv("S3_REGION")

	for {
		creds := credentials.NewStaticCredentials(accessKey, secretKey, "")
		_, err := creds.Get()
		if err != nil {
			logrus.Errorf("bad credentials: %s", err)
		}
		cfg := aws.NewConfig().WithEndpoint(endpointStr).WithDisableSSL(true).WithCredentials(creds).WithS3ForcePathStyle(true)
		if len(regionStr) > 0 {
			cfg = cfg.WithRegion(regionStr)
		} else {
			cfg = cfg.WithRegion(DefaultRegion)
		}

		svc := s3.New(session.New(), cfg)
		objName := uuid.New().String()
		fmt.Printf("--- PUT OBJECT %s IN BUCKET %s ---\n", objName, bucketName)
		putObject(svc, bucketName, objName)
		fmt.Printf("--- GET OBJECT %s IN BUCKET %s ---\n", objName, bucketName)
		getObject(svc, bucketName, objName)
		fmt.Printf("--- DONE ---\n")
		fmt.Printf("\n\n")

		creds = credentials.NewStaticCredentials(accessKey2, secretKey2, "")
		_, err = creds.Get()
		if err != nil {
			logrus.Errorf("bad credentials: %s", err)
		}
		cfg = aws.NewConfig().WithEndpoint(endpointStr).WithDisableSSL(true).WithCredentials(creds).WithS3ForcePathStyle(true)
		if len(regionStr) > 0 {
			cfg = cfg.WithRegion(regionStr)
		} else {
			cfg = cfg.WithRegion(DefaultRegion)
		}
		svc = s3.New(session.New(), cfg)
		objName = uuid.New().String()
		fmt.Printf("--- PUT OBJECT %s IN BUCKET %s ---\n", objName, bucketNameTwo)
		putObject(svc, bucketNameTwo, objName)
		fmt.Printf("--- GET OBJECT %s IN BUCKET %s ---\n", objName, bucketNameTwo)
		getObject(svc, bucketNameTwo, objName)
		fmt.Printf("--- DONE ---\n")
		fmt.Printf("\n\n")

		time.Sleep(15 * time.Second)
	}
}

func listBuckets(svc *s3.S3) {

	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")
	for _, b := range result.Buckets {
		fmt.Printf("* %s \n", aws.StringValue(b.Name))
	}
}

func createBucket(svc *s3.S3, bucketName string) {

	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	result, err := svc.CreateBucket(input)
	if handleError(err) {
		return
	}
	fmt.Println("Bucket created ", result)
}

func deleteBucket(svc *s3.S3, bucketName string) {

	input := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	}

	result, err := svc.DeleteBucket(input)
	if handleError(err) {
		return
	}
	fmt.Println("Bucket deleted ", result)
}

func putObject(svc *s3.S3, bucketName string, objName string) {

	b := bytes.NewBufferString("This string is just for testing")

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objName),
		Body:   bytes.NewReader(b.Bytes()),
	}

	result, err := svc.PutObject(input)
	if handleError(err) {
		return
	}
	fmt.Println(result)
}

func getObject(svc *s3.S3, bucketName string, objName string) {

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objName),
	}

	result, err := svc.GetObject(input)
	if handleError(err) {
		return
	}

	fmt.Println(result)
}

func deleteObject(svc *s3.S3, bucketName string, objName string) {

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objName),
	}

	result, err := svc.DeleteObject(input)
	if handleError(err) {
		return
	}

	fmt.Println(result)
}

func handleError(err error) bool {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return true
	}
	return false
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}
