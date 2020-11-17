package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io/ioutil"
)


// User is a retrieved and authenticated user.
type GoogleUser struct {
	Sub string `json:"sub"`
	Name string `json:"name"`
	GivenName string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Profile string `json:"profile"`
	Picture string `json:"picture"`
	Email string `json:"email"`
	EmailVerified bool `json:"email_verified"`
	Gender string `json:"gender"`
}


func userInfoFromToken (token *oauth2.Token) (err error, user *GoogleUser) {
	if !token.Valid() {
		log.Error(err)
		return
	}

	client := oauthConf.Client(oauth2.NoContext, token)
	response, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Error("Error getting user from token ", err)
		return
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	err = json.Unmarshal(contents, &user)
	if err != nil {
		log.Error("Error unmarshaling Google user", err)
		return
	}

	return
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

func getListBucketUser(userEmail string) ([]string) {

    var buckets []string
    for _,rule := range authRules.AuthRules {
        for _,user := range rule.Emails {
            if user == userEmail {
                buckets = append (buckets, rule.Buckets...)
            }
        }
    }

    uniqueBuckets := removeDuplicateStrings(buckets)

    return uniqueBuckets
}


func checkUserAuth(userEmail string) (bool) {

    for _,rule := range authRules.AuthRules {
        for _,user := range rule.Emails {
            if user == userEmail {
                log.Info("User is allowed ", user)
                return true
            }
        }
    }

    return false
}


func checkUserAuthBucket(userEmail string, userBucket string) (bool) {

    for _,rule := range authRules.AuthRules {
        for _,user := range rule.Emails {
            if user == userEmail {
                for _,bucket := range rule.Buckets {
                    if bucket == userBucket {
                        log.Info("User is allowed ", user, " to bucket ", bucket)
                        return true
                    }
                }
            }
        }
    }

    return false
}

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