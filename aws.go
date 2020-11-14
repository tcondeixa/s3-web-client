package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"time"
)

func AwsSessionCreate(profile, region string) (error, *session.Session) {

	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: profile,
		Config: aws.Config{
			Region: aws.String(region),
		},
		SharedConfigState: session.SharedConfigEnable,
	})

	return err, sess
}

func AwsS3BucketList(sess *session.Session, bucketName string) (error, []string) {

	var objectsList []string

	svc := s3.New(sess)
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: &bucketName,
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
		for _, obj := range p.Contents {
			objectsList = append(objectsList, *obj.Key)
		}
		return !last
	})
	if err != nil {
		return err, objectsList
	}

	return err, objectsList
}

func AwsS3BucketGet(sess *session.Session, bucketName string, bucketKey string) (error, []byte) {

	buf := aws.NewWriteAtBuffer([]byte{})

	downloader := s3manager.NewDownloader(sess)
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(bucketKey),
	})
	if err != nil {
		return err, []byte{}
	}

	return err, buf.Bytes()
}

func AwsS3PresignObjectGet(sess *session.Session, bucketName string, bucketKey string) (error, string) {

    svc := s3.New(sess)

    req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: &bucketName,
        Key:    &bucketKey,
    })

    urlStr, err := req.Presign(15 * time.Minute)
    if err != nil {
        return err, ""
    }

    return err, urlStr

}

func AwsCheckBucketExist(sess *session.Session, bucketName string) (error, bool) {

    svc := s3.New(sess)

    _, err := svc.HeadBucket(&s3.HeadBucketInput{
        Bucket: &bucketName,
    })

    if err != nil {
        aerr, ok := err.(awserr.Error)
        if (ok && aerr.Code() == s3.ErrCodeNoSuchBucket || aerr.Code() == "NotFound") {
            return nil, false
        }

        return err, false
    }

    return err, true
}


func checkAllBuckets(sess *session.Session, buckets []string) (error, []string) {

   var verifiedBuckets []string
   var bucketProblems error
   for _,bucket := range buckets {
        err, exist := AwsCheckBucketExist(sess, bucket)
        if err != nil {
            bucketProblems = err
            continue
        }

        if exist == true {
            verifiedBuckets = append(verifiedBuckets, bucket)
        }
   }

   return bucketProblems, verifiedBuckets
}