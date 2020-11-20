package main

import (
  	//"errors"
    "sort"
    //"github.com/aws/aws-sdk-go/aws/session"
    //"cloud.google.com/go/storage"
)


func checkBucketsConfigProviders (authRules []AuthRule) (bool, bool) {

    anyAwsBuket := false
    anyGcpBuket := false

    for _,rule := range authRules {
        if len(rule.AwsBuckets) > 0 {
            anyAwsBuket = true
        }

        if len(rule.GcpBuckets) > 0 {
            anyGcpBuket = true
        }

        if anyAwsBuket && anyGcpBuket {
            break
        }
    }

    return anyAwsBuket, anyGcpBuket
}


func sortAndValidateAuthRules (authRules []AuthRule) (error) {

    //var sess *session.Session
    //var client *storage.Client
    var err error

/*     hasAws, hasGcp := checkBucketsConfigProviders(authRules)

    if hasAws {
        err, sess = AwsSessionCreate()
        if err != nil {
            return err
        }
    }

    if hasGcp {
        err, client = GcpSessionCreate()
        if err != nil {
            return err
        }
    } */

    for _,rule := range authRules {
        sort.Strings(rule.Emails)

        if len(rule.AwsBuckets) > 0 {
            sort.Strings(rule.AwsBuckets)
/*             err, _ := checkAllAwsBuckets(sess, rule.AwsBuckets)
            if err != nil {
                err = errors.New("Please check if all AWS buckets in auth_rules exists and app has access")
                return err
            } */
        }

        if len(rule.GcpBuckets) > 0 {
            sort.Strings(rule.GcpBuckets)
/*             err, _ := checkAllGcpBuckets(client, rule.AwsBuckets)
            if err != nil {
                err = errors.New("Please check if all GCP buckets in auth_rules exists and app has access")
                return err
            } */
        }
    }

    return err
}