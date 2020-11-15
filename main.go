package main

import (
  	"context"
  	"flag"
  	"github.com/gorilla/mux"
  	"github.com/kelseyhightower/envconfig"
  	log "github.com/sirupsen/logrus"
  	"golang.org/x/oauth2"
  	"golang.org/x/oauth2/google"
  	"github.com/gorilla/securecookie"
  	"github.com/gorilla/sessions"
  	"net/http"
  	"os"
  	"fmt"
  	"os/signal"
  	"time"
  	"encoding/json"
  	"io/ioutil"
  	"sort"
  	"crypto/rand"
)


var config EnvVars
var authRules AuthRules
var oauthConf *oauth2.Config
var secureCookie *securecookie.SecureCookie
var store *sessions.CookieStore
var sessionTokenName string


type EnvVars struct {
	Host			string `default:"0.0.0.0" envconfig:"HOST"`
	Port			string `default:"8080" envconfig:"PORT"`
	Log     		string `default:"Info"  envconfig:"LOG_LEVEL"`
	Title	        string `default:"S3 Web Service" envconfig:"TITLE"`
	ClientID    	string `required:"true" envconfig:"CLIENT_ID"`
	ClientSecret	string `required:"true" envconfig:"CLIENT_SECRET"`
	RedirectURL     string `required:"true" envconfig:"REDIRECT_URL"`
	AuthFile        string `required:"true" envconfig:"AUTH_FILE"`
}

type AuthRules struct {
	AuthRules []AuthRule `json:"auth_rules"`
	BucketNames []BucketNaming `json:"bucket_friendly_naming"`
}

type AuthRule struct {
	Emails []string `json:"emails"`
	Buckets []string `json:"buckets"`
}

type BucketNaming struct {
	RealName string `json:"real_name"`
	FriendlyName string `json:"friendly_name"`
}


func logInit() {
	log.SetFormatter(&log.JSONFormatter{
		DisableHTMLEscape: true,
		PrettyPrint:       true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetReportCaller(true)
}

func authInit (ClientID, ClientSecret, RedirectUrl string) {
	oauthConf = &oauth2.Config{
		ClientID:     ClientID,
		ClientSecret: ClientSecret,
		RedirectURL:  RedirectUrl,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func sortAndValidateAuthRules (authRules []AuthRule) (bool) {

    err, sess := AwsSessionCreate("terraform","eu-central-1")
    if err != nil {
        log.Error(err)
        return false
    }

    for _,rule := range authRules {
        sort.Strings(rule.Emails)
        sort.Strings(rule.Buckets)
        err, _ := checkAllBuckets(sess, rule.Buckets)
        if err != nil {
            log.Error("Please check if all buckets in auth_rules exists and app has access")
            return false
        }
    }

    return true
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}

	return string(bytes), nil
}


func main() {

	logInit()

	err := envconfig.Process("", &config)
	if err != nil {
		log.Error("Fail to parse Env variables", err)
		os.Exit(1)
	}

	level, err := log.ParseLevel(config.Log)
	if err != nil {
        log.Error(err)
        os.Exit(1)
	}

    log.SetLevel(level)
    log.Info("Log level set to ", level)

    cookiesHashKey, err := GenerateRandomString(64)
    if err != nil {
        log.Error(err)
        os.Exit(1)
    }

	jsonFile, err := os.Open(config.AuthFile)
	defer jsonFile.Close()
    if err != nil {
        log.Error("Fail to parse Auth file", err)
        os.Exit(1)
    }

    byteValue, _ := ioutil.ReadAll(jsonFile)
    json.Unmarshal(byteValue, &authRules)
    valid := sortAndValidateAuthRules(authRules.AuthRules)
    if !valid {
        log.Error("Fail to reach buckets", err)
        os.Exit(1)
    }

    log.Info(authRules)
    sessionTokenName = "s3-web-client-token"

	authInit(config.ClientID, config.ClientSecret, config.RedirectURL)
	secureCookie = securecookie.New([]byte(cookiesHashKey), nil)
	store = sessions.NewCookieStore([]byte(cookiesHashKey))
	store.Options = &sessions.Options{
		MaxAge:   60 * 15, // 15 min
		HttpOnly: true,
	}

    var wait time.Duration
    flag.DurationVar(&wait, "graceful-timeout", time.Second*30, "the duration for which the server gracefully wait for existing connections to finish")
    flag.Parse()

    r := mux.NewRouter()
    r.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.Dir("./static/css"))))
    r.PathPrefix("/js").Handler(http.StripPrefix("/js", http.FileServer(http.Dir("./static/js"))))
    r.HandleFunc("/login", loginHandler).Methods("GET")
    r.HandleFunc("/logout", logoutHandler).Methods("GET")
    r.HandleFunc("/auth", authHandler).Methods("GET")
    r.HandleFunc("/main/{bucket}", bucketHandler).Methods("GET")
    r.HandleFunc("/health", healthHandler).Methods("GET")

    log.Info("Starting Server at port ", config.Port)
    srv := &http.Server{
        Addr: fmt.Sprintf("%s:%s", config.Host, config.Port),
        WriteTimeout: time.Second * 15,
        ReadTimeout:  time.Second * 15,
        IdleTimeout:  time.Second * 60,
        Handler:      r,
    }

    go func() {
        if err := srv.ListenAndServe(); err != nil {
            log.Error(err)
        }
    }()

    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)

    // Block until we receive our signal.
    <-c

    // Create a deadline to wait for.
    ctx, cancel := context.WithTimeout(context.Background(), wait)
    defer cancel()
    srv.Shutdown(ctx)

    log.Info("shutting down")
    os.Exit(0)
}
