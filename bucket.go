package main

import (
	log "github.com/sirupsen/logrus"
	"sort"
	"regexp"
)

func getVerifiedBucket (userEmail string) (error, string) {

    err, allowedBuckets := getListBucketUserMatching(getListBucketUserConfig(userEmail))
    if err != nil {
        log.Error(err)
        return err, ""
    }

    for _,bucket := range allowedBuckets {
        provider := getBucketProvider(bucket)

        exist := false
        if provider == "aws" {
            err, sess := AwsSessionCreate()
            if err != nil {
                log.Error(err)
                continue
            }

            err, exist = AwsCheckBucketExist(sess, bucket)
            if err != nil {
                log.Error(err)
                continue
            }

        } else if provider == "gcp" {
            err, client := GcpSessionCreate()
            if err != nil {
                log.Error(err)
                continue
            }

            err, exist = GcpCheckBucketExist(client, bucket)
            if err != nil {
                log.Error(err)
                continue
            }

        } else {
            log.Error("Unknown provider ", provider)
            continue
        }

        if exist == true {
            return nil, bucket
        }
    }

    return nil, ""
}

func orderBuckets (selectBucket string, buckets []string) ([]string) {

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

    return buckets
}

func getListBucketUserMatching (bucketsAws []string, bucketsGcp []string) (error, []string) {

    var matchBuckets []string
    //ListBucketsAws
    //Compare if match with regex and append to a new slice
    if len(bucketsAws) > 0 {
        err, sess := AwsSessionCreate()
        if err != nil {
            log.Error(err)
            return err, nil
        }

        err, allBuckets := AwsS3ListBuckets(sess)
        if err != nil {
            log.Error(err)
            return err, nil
        }

        for _, bucketRemote := range allBuckets {
            for _, bucketLocal := range bucketsAws {
                found, err := regexp.MatchString(bucketLocal, bucketRemote)
                if err != nil {
                    log.Error(err)
                    continue
                }

                if found {
                    matchBuckets = append(matchBuckets, bucketRemote)
                }
            }
        }
    }

    //ListBucketsGcp
    //Compare if match with regex and append to a new slice
    if len(bucketsGcp) > 0 {
        err, client := GcpSessionCreate()
        if err != nil {
            log.Error(err)
            return err, nil
        }

        err, allBuckets := GcpListBuckets(client)
        if err != nil {
            log.Error(err)
            return err, nil
        }

        for _, bucketRemote := range allBuckets {
            for _, bucketLocal := range bucketsAws {
                found, err := regexp.MatchString(bucketLocal, bucketRemote)
                if err != nil {
                    log.Error(err)
                    continue
                }

                if found {
                    matchBuckets = append(matchBuckets, bucketRemote)
                }
            }
        }
    }

    return nil, matchBuckets
}