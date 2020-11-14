package main

import (
	"encoding/json"
	"fmt"
	"github.com/dchest/uniuri"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"html/template"
	"net/http"
	"github.com/gorilla/mux"
)


type replyObjects struct {
	Email string
	Picture string
	S3Buckets []string
	S3Bucket string
	S3Objects []string
}


type signedUrl struct {
	Url string
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	oauthStateString := uniuri.New()
	url := oauthConf.AuthCodeURL(oauthStateString)

	tmpl := template.Must(template.ParseFiles("templates/login.tmpl"))
	templateData := map[string]interface{}{
		"link": url,
	}
	tmpl.ExecuteTemplate(w, "login.tmpl", templateData)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")

	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Error(err)
		return
	}

	if !token.Valid() {
		fmt.Fprintf(w, "Fail on Oauth authentication")
		return
	}

	err, user := userInfoFromToken(token)
	if err != nil || user.EmailVerified == false {
		log.Error(err)
        http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

    isAllowed := checkUserAuth(user.Email)
    if isAllowed == false {
        http.Redirect(w, r, "/login", http.StatusFound)
    }

	session, _ := store.Get(r, sessionTokenName)
	session.Values["oauth"] = token.AccessToken
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

    err, sess := AwsSessionCreate("terraform","eu-central-1")
    if err != nil {
        return
    }

    var verifiedBucket string
    allowedBuckets := getListBucketUser(user.Email)
    for _,bucket := range allowedBuckets {
        err, exist := AwsCheckBucketExist(sess, bucket)
        if err != nil {
            log.Error(err)
            continue
        }

        if exist == true {
            verifiedBucket = bucket
            break
        }
    }

    redirect := "/main/"+verifiedBucket
    http.Redirect(w, r, redirect, http.StatusFound)
}

func bucketHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, sessionTokenName)
	if err != nil {
		log.Error(err)
	}

	token := oauth2.Token{
		AccessToken: fmt.Sprintf("%v", session.Values["oauth"]),
	}

	if !token.Valid() {
		log.Error("Failure in Token Validation")
        http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	err, user := userInfoFromToken(&token)
	if err != nil || user.EmailVerified == false {
		log.Error(err)
        http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

    err, sess := AwsSessionCreate("terraform","eu-central-1")
    if err != nil {
        return
    }

    var verifiedBuckets []string
    allowedBuckets := getListBucketUser(user.Email)
    for _,bucket := range allowedBuckets {
        err, exist := AwsCheckBucketExist(sess, bucket)
        if err != nil {
            log.Error(err)
            continue
        }

        if exist == true {
            verifiedBuckets = append(verifiedBuckets, bucket)
        }
    }

    vars := mux.Vars(r)
    s3Bucket := getRealBucketName(vars["bucket"])

    selectedBucketPos := 0
    firstBucketValue := verifiedBuckets[0]
    for index,bucket := range verifiedBuckets {
        if bucket == s3Bucket {
            selectedBucketPos = index
            break
        }
    }
    verifiedBuckets[0] = s3Bucket
    verifiedBuckets[selectedBucketPos] = firstBucketValue


    isAllowed := checkUserAuthBucket(user.Email,s3Bucket)
    if isAllowed == false {
        http.Redirect(w, r, "/main", http.StatusFound)
    }

    s3Object := r.URL.Query().Get("object")
    if s3Object != "" {
        err, presignUrl := AwsS3PresignObjectGet(sess, s3Bucket, s3Object)
        if err != nil {
            return
        }

        http.Redirect(w, r, presignUrl, http.StatusFound)
    }

    err, objectsList := AwsS3BucketList(sess, s3Bucket)
    if err != nil {
        return
    }

    s3Bucket = getFriendlyBucketName(s3Bucket)
    verifiedBuckets = changeRealToFriendlyBuckets(verifiedBuckets)
    tmpl := template.Must(template.ParseFiles("templates/bucket.tmpl"))
    templateData := replyObjects {
        Email: user.Email,
        Picture: user.Picture,
        S3Buckets: verifiedBuckets,
        S3Bucket: s3Bucket,
        S3Objects: objectsList,
    }

    tmpl.ExecuteTemplate(w, "bucket.tmpl", templateData)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionTokenName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["testing"] = ""
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode("All Always Ok for Now, need to be improved")
}
