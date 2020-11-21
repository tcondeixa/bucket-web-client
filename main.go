package main

import (
  	"context"
  	"flag"
  	"github.com/gorilla/mux"
  	"github.com/kelseyhightower/envconfig"
  	log "github.com/sirupsen/logrus"
  	"golang.org/x/oauth2"
  	"github.com/gorilla/sessions"
  	"github.com/gorilla/securecookie"
  	"net/http"
  	"os"
  	"fmt"
  	"os/signal"
  	"time"
  	"encoding/json"
  	"io/ioutil"
)

var config EnvVars
var authRules AuthRules
var oauthConf *oauth2.Config
var store *sessions.CookieStore
var sessionTokenName string


func logInit() {
	log.SetFormatter(&log.JSONFormatter{
		DisableHTMLEscape: true,
		PrettyPrint:       true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetReportCaller(true)
}


func main() {

	logInit()

	err := envconfig.Process("", &config)
	if err != nil {
		log.Error("Fail to parse Env variables, ", err)
		os.Exit(1)
	}

	level, err := log.ParseLevel(config.Log)
	if err != nil {
        log.Error(err)
        os.Exit(1)
	}

    log.SetLevel(level)
    log.Info("Log level set to ", level)

	jsonFile, err := os.Open(config.AuthFile)
	defer jsonFile.Close()
    if err != nil {
        log.Error("Fail to parse Auth file ", err)
        os.Exit(1)
    }

    byteValue, _ := ioutil.ReadAll(jsonFile)
    json.Unmarshal(byteValue, &authRules)
    err = sortAndValidateAuthRules(authRules.AuthRules)
    if err != nil {
        log.Error(err)
        os.Exit(1)
    }

    log.Info(authRules)
    sessionTokenName = "s3-web-client-token"
	authInit(config.ClientID, config.ClientSecret, config.RedirectURL)
	store = sessions.NewCookieStore(securecookie.GenerateRandomKey(64),securecookie.GenerateRandomKey(32))
	store.Options = &sessions.Options{
		MaxAge:   60 * 60, // 1 hour to match google oauth token
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

    log.Info("Starting Server with host ",config.Host, " and port ", config.Port)
    srv := &http.Server{
        Addr: fmt.Sprintf("%s:%s", config.Host, config.Port),
        WriteTimeout: time.Second * 15,
        ReadTimeout:  time.Second * 15,
        IdleTimeout:  time.Second * 60,
        Handler:      r,
    }

    go func() {
        err := srv.ListenAndServe()
        if err != nil {
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
