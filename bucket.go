package main

import (
	log "github.com/sirupsen/logrus"
	"sort"
	"regexp"
	"sync"
	"time"
)

type BucketInfo struct {
	Name string
	Provider string
}

var signUrlValidMin time.Duration = 15


func orderBuckets (selectBucket string, buckets []string) ([]string) {

    log.Trace(selectBucket, buckets)

    // Order bucket slice but having the selected bucket in the first position
    selectedBucketPos := 0
    firstBucketValue := buckets[0]
    for index, bucket := range buckets {
        if bucket == selectBucket {
            selectedBucketPos = index
            break
        }
    }
    buckets[0] = selectBucket
    buckets[selectedBucketPos] = firstBucketValue
    sort.Strings(buckets[1:])

    log.Trace(buckets)

    return buckets
}


func getMatchedBucketUserAws (wg *sync.WaitGroup, matchBuckets *[]string, bucketsAws []string) {

    log.Trace(*matchBuckets, bucketsAws)
    defer wg.Done()

    err, allBuckets := AwsS3ListBuckets()
    if err != nil {
        log.Error(err)
        return
    }

    for _, bucketRemote := range allBuckets {
        for _, bucketLocal := range bucketsAws {
            found, err := regexp.MatchString(bucketLocal, bucketRemote)
            if err != nil {
                log.Error(err)
                continue
            }

            if found {
                *matchBuckets = append(*matchBuckets, bucketRemote)
            }
        }
    }

    log.Trace(*matchBuckets)
}


func getMatchedBucketUserGcp (wg *sync.WaitGroup, matchBuckets *[]string, bucketsGcp []string) {

    log.Trace(*matchBuckets, bucketsGcp)
    defer wg.Done()

    err, allBuckets := GcpListBuckets()
    if err != nil {
        log.Error(err)
        return
    }

    for _, bucketRemote := range allBuckets {
        for _, bucketLocal := range bucketsGcp {
            found, err := regexp.MatchString(bucketLocal, bucketRemote)
            if err != nil {
                log.Error(err)
                continue
            }

            if found {
                *matchBuckets = append(*matchBuckets, bucketRemote)
            }
        }
    }

    log.Trace(*matchBuckets)
}


func getListBucketUserMatching (bucketsAws, bucketsGcp []string) ([]BucketInfo) {

    log.Trace(bucketsAws, bucketsGcp)

    var wg sync.WaitGroup
    var matchAwsBuckets []string
    var matchGcpBuckets []string

    //ListBucketsAws
    if len(bucketsAws) > 0 {
        wg.Add(1)
        go getMatchedBucketUserAws(&wg, &matchAwsBuckets, bucketsAws)
    }

    //ListBucketsGcp
    if len(bucketsGcp) > 0 {
        wg.Add(1)
        go getMatchedBucketUserGcp(&wg, &matchGcpBuckets, bucketsGcp)
    }

    wg.Wait()
    log.Trace(matchAwsBuckets, matchGcpBuckets)

    matchBuckets := make([]BucketInfo, len(matchAwsBuckets)+len(matchGcpBuckets))
    for i,name := range matchAwsBuckets {
        matchBuckets[i].Name = name
        matchBuckets[i].Provider = "aws"
    }

    for i,name := range matchGcpBuckets {
        matchBuckets[i+len(matchAwsBuckets)].Name = name
        matchBuckets[i+len(matchAwsBuckets)].Provider = "gcp"
    }

    log.Trace(matchBuckets)
    return matchBuckets
}


func getSignedBucketUrl (bucketList []BucketInfo, bucket, object string) (string) {

    log.Trace(bucketList, bucket, object)

    var presignUrl string
    var err error

    for _,b := range bucketList {
        if b.Name == bucket {
            if b.Provider == "aws" {
                err, presignUrl = AwsS3PresignObjectGet(bucket, object)
                if err != nil {
                    log.Error(err)
                    return ""
                }

                return presignUrl

            } else if b.Provider == "gcp" {
                err, presignUrl = GcpPresignObjectGet(bucket, object)
                if err != nil {
                    log.Error(err)
                    return presignUrl
                }

                return presignUrl

            } else {
                log.Error("unknown provider to signedUrl ", b.Provider, b.Name)
                return ""
            }
        }
    }

    log.Trace(presignUrl)
    return ""
}

func getBucketObjectsList (bucketList []BucketInfo, bucket string) ([]string) {

    log.Trace(bucketList, bucket)

    var objectsList []string
    var err error

    for _,b := range bucketList {
        if b.Name == bucket {
            if b.Provider == "aws" {
                err, objectsList = AwsS3ListObjects(bucket)
                if err != nil {
                    log.Error(err)
                    return objectsList
                }

            } else if b.Provider == "gcp" {
                err, objectsList = GcpListObjects(bucket)
                if err != nil {
                    log.Error(err)
                    return objectsList
                }

            } else {
                log.Error("unknown provider to get bucket list ", b.Provider, b.Name)
                return objectsList
            }
        }
    }

    log.Trace(objectsList)
    return objectsList
}