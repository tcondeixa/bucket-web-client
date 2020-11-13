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
  	"strings"
)


var config EnvVars
var oauthConf *oauth2.Config
var secureCookie *securecookie.SecureCookie
var store *sessions.CookieStore

type MyStringList []string

func (values *MyStringList) Decode (input string) {
	*values = strings.Split(strings.TrimSpace(input), ",")
}

type EnvVars struct {
	Host			string `default:"0.0.0.0" envconfig:"HOST"`
	Port			string `default:"8080" envconfig:"PORT"`
	Service 		string `default:"Payrolling" envconfig:"SERVICE"`
	Log     		string `default:"Info"  envconfig:"LOG_LEVEL"`
	BucketName		string `required:"true" envconfig:"BUCKET_NAME"`
	AllowedEmails 	MyStringList `envconfig:"ALLOWED_EMAILS"`
	CookiesHashKey	string 	`required:"true" envconfig:"COOKIES_HASH_KEY"`
	ClientID    	string  `required:"true" envconfig:"CLIENT_ID"`
	ClientSecret	string  `required:"true" envconfig:"CLIENT_SECRET"`
	RedirectURL     string  `required:"true" envconfig:"REDIRECT_URL"`
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


func main() {

	logInit()

	err := envconfig.Process("", &config)
	if err != nil {
		log.Error("Fail to parse Env variables", err)
		os.Exit(1)
	}
	log.Info(config)

	level, err := log.ParseLevel(config.Log)
	if err != nil {
		log.SetLevel(level)
	}

	authInit(config.ClientID, config.ClientSecret, config.RedirectURL)
	secureCookie = securecookie.New([]byte(config.CookiesHashKey), nil)
	store = sessions.NewCookieStore([]byte(config.CookiesHashKey))
	store.Options = &sessions.Options{
		MaxAge:   60 * 15, // 15 min
		HttpOnly: true,
	}

    var wait time.Duration
    flag.DurationVar(&wait, "graceful-timeout", time.Second*30, "the duration for which the server gracefully wait for existing connections to finish")
    flag.Parse()

    r := mux.NewRouter()
    r.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.Dir("./static/css"))))
    r.PathPrefix("/img").Handler(http.StripPrefix("/img", http.FileServer(http.Dir("./static/img"))))
    r.HandleFunc("/login", loginHandler).Methods("GET")
    r.HandleFunc("/logout", logoutHandler).Methods("GET")
    r.HandleFunc("/auth", authHandler).Methods("GET")
    r.HandleFunc("/main", mainHandler).Methods("GET")
    r.HandleFunc("/download", downloadHandler).Methods("GET")


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
