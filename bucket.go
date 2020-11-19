package main

import (
	log "github.com/sirupsen/logrus"
	"sort"
)

func getVerifiedBucket (userEmail string) (error, string) {

    allowedBuckets := getListBucketUser(userEmail)
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