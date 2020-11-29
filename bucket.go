package main

import (
	log "github.com/sirupsen/logrus"
	"sort"
	"regexp"
	"time"
)

type BucketInfo struct {
	Name string
	Provider string
}

var signUrlValidMin time.Duration = 15


func orderStringSlice (selectElement string, elements []string) ([]string) {

    log.Trace(selectElement, elements)

    // Order slice but having the selected element in the first position
    selectElementPos := 0
    firstBucketValue := elements[0]
    for index, bucket := range elements {
        if bucket == selectElement {
            selectElementPos = index
            break
        }
    }
    elements[0] = selectElement
    elements[selectElementPos] = firstBucketValue
    sort.Strings(elements[1:])

    log.Trace(elements)

    return elements
}


func orderIntSlice (selectElement int, elements []int) ([]int) {

    log.Trace(selectElement, elements)

    // Order slice but having the selected element in the first position
    selectElementPos := 0
    firstBucketValue := elements[0]
    for index, bucket := range elements {
        if bucket == selectElement {
            selectElementPos = index
            break
        }
    }
    elements[0] = selectElement
    elements[selectElementPos] = firstBucketValue
    sort.Ints(elements[1:])

    log.Trace(elements)

    return elements
}


func getMatchedBucketUserAws (bucketsAws []string) ([]string) {

    log.Trace(bucketsAws)

    var matchBuckets []string

    semaphoreAws <- struct{}{}
    allBuckets := awsListBuckets
    <-semaphoreAws

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

    log.Trace(matchBuckets)
    return matchBuckets
}


func getMatchedBucketUserGcp (bucketsGcp []string) ([]string) {

    log.Trace(bucketsGcp)

    var matchBuckets []string

    semaphoreGcp <- struct{}{}
    allBuckets := gcpListBuckets
    <-semaphoreGcp

    for _, bucketRemote := range allBuckets {
        for _, bucketLocal := range bucketsGcp {
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

    log.Trace(matchBuckets)
    return matchBuckets
}


func getAllBuckets () {

    err, allBuckets := AwsS3ListBuckets()
    if err != nil {
        log.Error(err)
        return
    }

    semaphoreAws <- struct{}{}
    awsListBuckets = allBuckets
    <-semaphoreAws
    log.Info(allBuckets)

    err, allBuckets = GcpListBuckets()
    if err != nil {
        log.Error(err)
        return
    }

    semaphoreGcp <- struct{}{}
    gcpListBuckets = allBuckets
    <-semaphoreGcp
    log.Info(allBuckets)
}


func getListBucketUserMatching (bucketsAws, bucketsGcp []string) ([]BucketInfo) {

    log.Trace(bucketsAws, bucketsGcp)

    var matchAwsBuckets []string
    var matchGcpBuckets []string

    //ListBucketsAws
    if len(bucketsAws) > 0 {
        matchAwsBuckets = getMatchedBucketUserAws(bucketsAws)
    }

    //ListBucketsGcp
    if len(bucketsGcp) > 0 {
        matchGcpBuckets = getMatchedBucketUserGcp(bucketsGcp)
    }

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

func getBucketObjectsList (bucketList []BucketInfo, bucket string, order string) ([]string) {

    log.Trace(bucketList, bucket)

    var objectsList []string

    for _,b := range bucketList {
        if b.Name == bucket {
            objectsList = getBucketObjectsCache(bucket,b.Provider)
        }
    }

    if order == "zA" {
        sort.Sort(sort.Reverse(sort.StringSlice(objectsList)))
    }

    log.Trace(objectsList)
    return objectsList
}