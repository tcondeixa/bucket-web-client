package main

import (
	"time"
	"io/ioutil"
	"context"

	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"cloud.google.com/go/storage"
    "google.golang.org/api/cloudresourcemanager/v1"
)


func GcpSessionCreate() (error, *storage.Client) {

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err, nil
	}

	return err, client
}


func GcpListBuckets(client *storage.Client) (error, []string) {

	var bucketsList []string

    ctx := context.Background()
    c, err := google.DefaultClient(ctx, cloudresourcemanager.CloudPlatformScope)
    if err != nil {
        return err, nil
    }

    cloudresourcemanagerService, err := cloudresourcemanager.New(c)
    if err != nil {
        return err, nil
    }

    req := cloudresourcemanagerService.Projects.List()
    err = req.Pages(ctx,func(page *cloudresourcemanager.ListProjectsResponse) error {
        for _, project := range page.Projects {
            fmt.Printf("%v\n", project)

            it := client.Buckets(ctx, project.ProjectId)
            for {
                attrs, err := it.Next()
                if err == iterator.Done {
                        break
                }

                if err != nil {
                        return err
                }

                bucketsList = append(bucketsList, attrs.Name)
            }
        }

        return nil

    });

    if err != nil {
        return err, nil
    }

	return nil, bucketsList
}


func GcpListObjects(client *storage.Client, bucketName string) (error, []string) {

	var objectsList []string

    ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*30)
	defer cancel()

	it := client.Bucket(bucketName).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return err, nil
		}

		objectsList = append(objectsList, attrs.Name)
	}

	return nil, objectsList
}


func GcpPresignObjectGet(client *storage.Client, bucketName, objectName string) (error, string) {

	jsonKey, err := ioutil.ReadFile(config.GoogleFile)
	if err != nil {
		return err, ""
	}

	conf, err := google.JWTConfigFromJSON(jsonKey)
	if err != nil {
		return err, ""
	}

	url, err := storage.SignedURL(bucketName, objectName, &storage.SignedURLOptions{
        Scheme:         storage.SigningSchemeV4,
        Method:         "GET",
        GoogleAccessID: conf.Email,
        PrivateKey:     conf.PrivateKey,
        Expires:        time.Now().Add(15 * time.Minute),
    })
	if err != nil {
		return err, ""
	}

	return nil, url
}

func GcpCheckBucketExist(client *storage.Client, bucketName string) (error, bool) {

    ctx := context.Background()
    ctx, cancel := context.WithTimeout(ctx, time.Second*10)
    defer cancel()

	_, err := client.Bucket(bucketName).Attrs(ctx)
	if err != nil {
        if err == storage.ErrBucketNotExist {
            return nil, false
        }

		return err, false
	}

	return err, true
}

func checkAllGcpBuckets(sess *storage.Client, buckets []string) (error, []string) {

   var verifiedBuckets []string
   var bucketProblems error
   for _,bucket := range buckets {
        err, exist := GcpCheckBucketExist(sess, bucket)
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