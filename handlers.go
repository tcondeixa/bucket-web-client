package main

import (
	"encoding/json"
	"fmt"
	"github.com/dchest/uniuri"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"html/template"
	"io/ioutil"
	"net/http"
)



type replyUser struct {
	Email string
	Picture string
	S3Objects []string
}

type signedUrl struct {
	Url string
}

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
	if err != nil {
		log.Error(err)
		return
	}
	log.Info(user)

	session, _ := store.Get(r, "talent-payrolling-s3-token")
	session.Values["oauth"] = token.AccessToken
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/main", http.StatusFound)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "talent-payrolling-s3-token")
	if err != nil {
		log.Error(err)
	}
	log.Info(session.Values["oauth"])


	// Validate the Token
	token := oauth2.Token{
		AccessToken: fmt.Sprintf("%v", session.Values["oauth"]),
	}

	if !token.Valid() {
		log.Error("Failure in Authentication Middleware")
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Not Authenticated User")
		return
	}

	err, user := userInfoFromToken(&token)
	if err != nil || user.EmailVerified == false {
		log.Error(err)
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Not Authenticated User")
		return
	}

    isAdmin := false
    for _, admin := range config.AllowedEmails {
        if admin == user.Email {
            log.Info("Admin is accessing")
            isAdmin = true
            break
        }
    }

    if isAdmin == false {
        http.Redirect(w, r, "/login", http.StatusFound)
    }

    err, sess := AwsSessionCreate("terraform","eu-central-1")
    if err != nil {
        return
    }

    err, objectsList := AwsS3BucketList(sess, config.BucketName)
    if err != nil {
        return
    }

	tmpl := template.Must(template.ParseFiles("templates/main.tmpl"))
	templateData := replyUser {
		Email: user.Email,
		Picture: user.Picture,
		S3Objects: objectsList,
	}
	tmpl.ExecuteTemplate(w, "main.tmpl", templateData)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, "talent-payrolling-s3-token")
	if err != nil {
		log.Error(err)
	}
	log.Info(session.Values["oauth"])


	// Validate the Token
	token := oauth2.Token{
		AccessToken: fmt.Sprintf("%v", session.Values["oauth"]),
	}

	if !token.Valid() {
		log.Error("Failure in Authentication Middleware")
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Not Authenticated User")
		return
	}

	err, user := userInfoFromToken(&token)
	if err != nil || user.EmailVerified == false {
		log.Error(err)
		w.WriteHeader(http.StatusForbidden)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Not Authenticated User")
		return
	}

    isAdmin := false
    for _, admin := range config.AllowedEmails {
        if admin == user.Email {
            log.Info("Admin is accessing")
            isAdmin = true
            break
        }
    }

    if isAdmin == false {
        http.Redirect(w, r, "/login", http.StatusFound)
    }


    err, sess := AwsSessionCreate("terraform","eu-central-1")
    if err != nil {
        return
    }

    s3Object := r.URL.Query().Get("object")
    log.Info(s3Object)

    err, presignUrl := AwsS3PresignObjectGet(sess, config.BucketName, s3Object)
    if err != nil {
        return
    }
    log.Info(presignUrl)

    http.Redirect(w, r, presignUrl, http.StatusFound)
}


func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "talent-payrolling-s3-token")
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

