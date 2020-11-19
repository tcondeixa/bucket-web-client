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
	Buckets []string
	Bucket string
	Objects []string
}


func loginHandler(w http.ResponseWriter, r *http.Request) {

	oauthStateString := uniuri.New()
	url := oauthConf.AuthCodeURL(oauthStateString)

	tmpl := template.Must(template.ParseFiles("templates/login.tmpl"))
	templateData := map[string]interface{}{
		"Link": url,
		"Title": config.Title,
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
        log.Info("unauthorised user trying to access ", user.Email)
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
    err, verifiedBucket := getVerifiedBucket(user.Email)
    if err != nil || verifiedBucket == "" {
        log.Error("No Buckets allowed for user ", user.Email)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
    bucket := getFriendlyBucketName(verifiedBucket)
    http.Redirect(w, r, "/main/"+bucket, http.StatusFound)
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
    allowedBuckets := getListBucketUser(user.Email)

    log.Trace("")
    vars := mux.Vars(r)
    bucket := getRealBucketName(vars["bucket"])
    provider := getBucketProvider(bucket)
    if provider != "aws" && provider != "gcp" {
        log.Error("Bucket does not belong to any provider")
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
    allowedBuckets = orderBuckets(bucket, allowedBuckets)
    isAllowed := checkUserAuthBucket(user.Email, bucket)
    if isAllowed == false {
        log.Info("unauthorised user trying to access", user.Email)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
    object := r.URL.Query().Get("object")
    if object != "" {

        var presignUrl string
        if provider == "aws" {
            err, sess := AwsSessionCreate()
            if err != nil {
                log.Error(err)
                http.Redirect(w, r, "/login", http.StatusFound)
                return
            }

            err, presignUrl = AwsS3PresignObjectGet(sess, bucket, object)
        } else if provider == "gcp" {
            err, client := GcpSessionCreate()
            if err != nil {
                log.Error(err)
                http.Redirect(w, r, "/login", http.StatusFound)
                return
            }

            err, presignUrl = GcpPresignObjectGet(client, bucket, object)
        }

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
    var objectsList []string
    if provider == "aws" {
        err, sess := AwsSessionCreate()
        if err != nil {
            log.Error(err)
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        err, objectsList = AwsS3BucketList(sess, bucket)
    } else if provider == "gcp" {
        err, client := GcpSessionCreate()
        if err != nil {
            log.Error(err)
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        err, objectsList = GcpBucketList(client, bucket)
    }

    if err != nil {
        log.Error(err)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    log.Trace("")
    tmpl := template.Must(template.ParseFiles("templates/bucket.tmpl"))
    log.Trace("")
    templateData := replyObjects {
        Title: config.Title,
        Email: user.Email,
        Picture: user.Picture,
        Buckets: changeRealToFriendlyBuckets(allowedBuckets),
        Bucket: getFriendlyBucketName(bucket),
        Objects: objectsList,
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

	session.Values["oauth"] = ""
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
