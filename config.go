package main

import (
    "sort"
    "regexp"
    log "github.com/sirupsen/logrus"
)

type EnvVars struct {
	Host			string `default:"0.0.0.0" envconfig:"HOST"`
	Port			string `default:"8080" envconfig:"PORT"`
	Log     		string `default:"Info"  envconfig:"LOG_LEVEL"`
	Title	        string `default:"Bucket Web Client" envconfig:"TITLE"`
	ClientID    	string `required:"true" envconfig:"CLIENT_ID"`
	ClientSecret	string `required:"true" envconfig:"CLIENT_SECRET"`
	RedirectURL     string `required:"true" envconfig:"REDIRECT_URL"`
	AuthFile        string `required:"true" envconfig:"AUTH_FILE"`
	GoogleFile      string `default:"" envconfig:"GOOGLE_APPLICATION_CREDENTIALS"`
}

type AuthRules struct {
	AuthRules []AuthRule `json:"auth_rules"`
	BucketNames []BucketNaming `json:"bucket_friendly_naming"`
}

type AuthRule struct {
	Emails []string `json:"emails"`
	AwsBuckets []string `json:"aws_buckets"`
	GcpBuckets []string `json:"gcp_buckets"`
}

type BucketNaming struct {
	RealName string `json:"real_name"`
	FriendlyName string `json:"friendly_name"`
}


func sortAndValidateAuthRules (authRules []AuthRule) (error) {

    for _,rule := range authRules {
        sort.Strings(rule.Emails)
        for _,r := range rule.Emails {
            _, err := regexp.Compile(r)
            if err != nil {
                return err
            }
        }

        if len(rule.AwsBuckets) > 0 {
            sort.Strings(rule.AwsBuckets)
            for _,r := range rule.AwsBuckets {
                _, err := regexp.Compile(r)
                if err != nil {
                    return err
                }
            }
        }

        if len(rule.GcpBuckets) > 0 {
            sort.Strings(rule.GcpBuckets)
            for _,r := range rule.GcpBuckets {
                _, err := regexp.Compile(r)
                if err != nil {
                    return err
                }
            }
        }
    }

    return nil
}

func removeDuplicateStrings(slice []string) ([]string) {
    keys := map[string]bool{}
    list := []string{}

    for _, entry := range slice {
        if _, value := keys[entry]; !value {
            keys[entry] = true
            list = append(list, entry)
        }
    }

    return list
}


func getListBucketUserConfig(userEmail string) ([]string, []string) {

    var bucketsAws, bucketsGcp []string
    for _,rule := range authRules.AuthRules {
        for _,user := range rule.Emails {
            found, err := regexp.MatchString(user, userEmail)
            if err != nil {
                log.Error(err)
                continue
            }

            if found {
                bucketsAws = append (bucketsAws, rule.AwsBuckets...)
                bucketsGcp = append (bucketsGcp, rule.GcpBuckets...)
            }
        }
    }

    uniqueBucketsAws := removeDuplicateStrings(bucketsAws)
    uniqueBucketsGcp := removeDuplicateStrings(bucketsGcp)

    return uniqueBucketsAws, uniqueBucketsGcp
}

func checkUserAuth(userEmail string) (bool) {

    for _,rule := range authRules.AuthRules {
        for _,user := range rule.Emails {
            found, err := regexp.MatchString(user, userEmail)
            if err != nil {
                log.Error(err)
                continue
            }

            if found {
                log.Info("User ", userEmail, " is allowed")
                return true
            }
        }
    }

    return false
}


func checkUserAuthBucket(userEmail, userBucket string) (bool) {

    for _,rule := range authRules.AuthRules {
        for _,user := range rule.Emails {
            found, err := regexp.MatchString(user, userEmail)
            if err != nil {
                log.Error(err)
                continue
            }

            if found {
                for _,bucket := range rule.AwsBuckets {
                    found, err := regexp.MatchString(bucket, userBucket)
                    if err != nil {
                        log.Error(err)
                        continue
                    }

                    if found {
                        log.Info("User ", userEmail, " accessing Aws bucket ", userBucket)
                        return true
                    }
                }

                for _,bucket := range rule.GcpBuckets {
                    found, err := regexp.MatchString(bucket, userBucket)
                    if err != nil {
                        log.Error(err)
                        continue
                    }

                    if found {
                        log.Info("User ", userEmail, " accessing Gcp bucket ", userBucket)
                        return true
                    }
                }
            }
        }
    }

    return false
}


// FriendlyName related functions
func getRealBucketName(friendlyName string) string {

    for _,bucket := range authRules.BucketNames {
        if bucket.FriendlyName == friendlyName {
            return bucket.RealName
        }
    }

    return friendlyName
}


func getFriendlyBucketName(realName string) string {

    for _,bucket := range authRules.BucketNames {
        if bucket.RealName == realName {
            return bucket.FriendlyName
        }
    }

    return realName
}


func changeRealToFriendlyBuckets (realName []string) []string {

    var buckets []string

    for _,bucket := range realName {
        buckets = append(buckets, getFriendlyBucketName(bucket))
    }

    return buckets
}