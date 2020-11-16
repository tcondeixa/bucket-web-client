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
	Title string
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
	tmpl.ExecuteTemplate(w, "login.tmpl", &templateData)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")

    log.Trace("")
	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Error(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

    log.Trace("")
	if !token.Valid() {
		log.Error("Fail on Oauth authentication")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

    log.Trace("")
	err, user := userInfoFromToken(token)
	if err != nil || user.EmailVerified == false {
		log.Error(err)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
	}

    log.Trace("")
    isAllowed := checkUserAuth(user.Email)
    if isAllowed == false {
        log.Info("unauthorised user trying to access", user.Email)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
	session, _ := store.Get(r, sessionTokenName)
	session.Values["oauth"] = token.AccessToken
	err = session.Save(r, w)
	if err != nil {
	    log.Error(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

    log.Trace("")
    err, sess := AwsSessionCreate()
    if err != nil {
        log.Error(err)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
    var verifiedBucket string
    allowedBuckets := getListBucketUser(user.Email)
    log.Trace("")
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

    log.Trace("")
    if verifiedBucket == "" {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    http.Redirect(w, r, "/main/"+verifiedBucket, http.StatusFound)
}

func bucketHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, sessionTokenName)
	if err != nil {
		log.Error(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

    log.Trace("")
	token := oauth2.Token{
		AccessToken: fmt.Sprintf("%v", session.Values["oauth"]),
	}

    log.Trace("")
	if !token.Valid() {
		log.Error("Failure in Token Validation")
        http.Redirect(w, r, "/login", http.StatusFound)
        return
	}

    log.Trace("")
	err, user := userInfoFromToken(&token)
	if err != nil || user.EmailVerified == false {
		log.Error(err)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
	}

    log.Trace("")
    err, sess := AwsSessionCreate()
    if err != nil {
        log.Error(err)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
    allowedBuckets := getListBucketUser(user.Email)

    log.Trace("")
    vars := mux.Vars(r)
    s3Bucket := getRealBucketName(vars["bucket"])

    selectedBucketPos := 0
    firstBucketValue := allowedBuckets[0]
    for index,bucket := range allowedBuckets {
        if bucket == s3Bucket {
            selectedBucketPos = index
            break
        }
    }
    allowedBuckets[0] = s3Bucket
    allowedBuckets[selectedBucketPos] = firstBucketValue

    log.Trace("")
    isAllowed := checkUserAuthBucket(user.Email,s3Bucket)
    if isAllowed == false {
        log.Info("unauthorised user trying to access", user.Email)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
    s3Object := r.URL.Query().Get("object")
    if s3Object != "" {
        err, presignUrl := AwsS3PresignObjectGet(sess, s3Bucket, s3Object)
        if err != nil {
            log.Error(err)
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        log.Trace("")
        http.Redirect(w, r, presignUrl, http.StatusFound)
        return
    }

    log.Trace("")
    err, objectsList := AwsS3BucketList(sess, s3Bucket)
    if err != nil {
        log.Error(err)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
    s3Bucket = getFriendlyBucketName(s3Bucket)
    allowedBuckets = changeRealToFriendlyBuckets(allowedBuckets)
    log.Trace("")
    tmpl := template.Must(template.ParseFiles("templates/bucket.tmpl"))
    log.Trace("")
    templateData := replyObjects {
        Title: config.Title,
        Email: user.Email,
        Picture: user.Picture,
        S3Buckets: allowedBuckets,
        S3Bucket: s3Bucket,
        S3Objects: objectsList,
    }

    tmpl.ExecuteTemplate(w, "bucket.tmpl", &templateData)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, sessionTokenName)
	if err != nil {
	    log.Error(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	session.Values["testing"] = ""
	session.Options.MaxAge = -1

	err = session.Save(r, w)
	if err != nil {
	    log.Error(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/login", http.StatusFound)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode("All Always Ok for Now, need to be improved")
}
