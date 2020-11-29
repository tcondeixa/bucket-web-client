package main

import (
	log "github.com/sirupsen/logrus"
	"time"
	"errors"
)

type BucketObjects struct {
    Bucket string
	Objects []string
	Timestamp int64
	Provider string
}

var semaphoreCachedBucketObjects chan struct{}
var cacheBucketObjects []BucketObjects

func getBucketObjectsCache (bucket string, provider string) ([]string){

    var objectsList []string
    var err error

    found := false
    timestampNow := time.Now().Unix()

    semaphoreCachedBucketObjects <- struct{}{}
    for _,v := range cacheBucketObjects {
        if v.Bucket == bucket {
            if v.Timestamp + config.TimeoutCache < timestampNow {
                err, objectsList = getBucketObjects (bucket, provider)
                if err != nil {
                    log.Error(err)
                    objectsList = v.Objects
                } else {
                    v.Objects = objectsList
                    v.Timestamp = timestampNow
                }

            } else {
                objectsList = v.Objects
            }

            found = true
            break
        }
    }
    <-semaphoreCachedBucketObjects

    if found {
        return objectsList
    }

    err, objectsList = getBucketObjects (bucket, provider)
    if err != nil {
        log.Error(err)
        return objectsList
    }

    timestampNow = time.Now().Unix()
    newBucket := BucketObjects{
        Bucket: bucket,
        Objects: objectsList,
        Timestamp: timestampNow,
        Provider: provider,
    }

    semaphoreCachedBucketObjects <- struct{}{}
    cacheBucketObjects = append(cacheBucketObjects,newBucket)
    <-semaphoreCachedBucketObjects

    return objectsList
}


func getBucketObjects (bucket string, provider string) (error, []string){

    var objectsList []string
    var err error

    if provider == "aws" {
        err, objectsList = AwsS3ListObjects(bucket)

    } else if provider == "gcp" {
        err, objectsList = GcpListObjects(bucket)

    } else {
        err = errors.New("unknown provider "+provider+" to update bucketObjects "+bucket)
    }

    return err, objectsList
}