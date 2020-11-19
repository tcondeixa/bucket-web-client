package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"time"
)

func AwsSessionCreate() (error, *session.Session) {

	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})

	return err, sess
}

func AwsS3BucketList(sess *session.Session, bucketName string) (error, []string) {

	var objectsList []string

	svc := s3.New(sess)
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}, func(p *s3.ListObjectsOutput, last bool) (shouldContinue bool) {
	    objectsListPage := make([]string, len(p.Contents))
		for i, obj := range p.Contents {
			objectsListPage[i] = *obj.Key
		}
		objectsList = append(objectsList, objectsListPage...)
		return !last
	})
	if err != nil {
		return err, objectsList
	}

	return err, objectsList
}

func AwsS3PresignObjectGet(sess *session.Session, bucketName string, bucketKey string) (error, string) {

    svc := s3.New(sess)

    req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(bucketKey),
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
        Bucket: aws.String(bucketName),
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


func checkAllAwsBuckets(sess *session.Session, buckets []string) (error, []string) {

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