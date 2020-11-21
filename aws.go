package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"time"
)


func AwsS3ListBuckets() (error, []string) {

    sess, err := session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    })
	svc := s3.New(sess)
	req, err := svc.ListBuckets(nil)
	if err != nil {
		return err, nil
	}

    bucketsList := make([]string, len(req.Buckets))
    for i, b := range req.Buckets {
        bucketsList[i] = *b.Name
    }

	return err, bucketsList
}


func AwsS3ListObjects(bucketName string) (error, []string) {

	var objectsList []string

    sess, err := session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    })
	svc := s3.New(sess)

	err = svc.ListObjectsV2Pages(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}, func(p *s3.ListObjectsV2Output, last bool) (shouldContinue bool) {
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

func AwsS3PresignObjectGet(bucketName, bucketKey string) (error, string) {

    sess, err := session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    })
    svc := s3.New(sess)

    req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(bucketKey),
    })

    urlStr, err := req.Presign(signUrlValidMin * time.Minute)
    if err != nil {
        return err, ""
    }

    return err, urlStr

}

func AwsCheckBucketExist(bucketName string) (error, bool) {

    sess, err := session.NewSessionWithOptions(session.Options{
        SharedConfigState: session.SharedConfigEnable,
    })
    svc := s3.New(sess)

    _, err = svc.HeadBucket(&s3.HeadBucketInput{
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


func checkAllAwsBuckets(buckets []string) (error, []string) {

   var verifiedBuckets []string
   var bucketProblems error
   for _,bucket := range buckets {
        err, exist := AwsCheckBucketExist(bucket)
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