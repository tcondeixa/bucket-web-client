package main

import (
	log "github.com/sirupsen/logrus"
	"sort"
	"regexp"
	"sync"
	"time"
)

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


func getListBucketUserMatching (bucketsAws, bucketsGcp []string) ([]string) {

    log.Trace(bucketsAws, bucketsGcp)

    var wg sync.WaitGroup
    var matchBuckets []string
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
    matchBuckets = append(matchAwsBuckets, matchGcpBuckets...)
    return matchBuckets
}