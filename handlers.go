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
	"strconv"
)


type replyObjects struct {
	Title string
	Email string
	Picture string
	Buckets []string
	Bucket string
	Objects []string
	FilesPage []int
	FilesOrder []string
	Pages []int
	CurrentPage int
}

var objectsPerPageOptions []int = []int{25, 50, 100}
var objectsOrderOptions []string = []string{"aZ","zA"}


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

	token, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Error(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	if !token.Valid() {
		log.Error("Fail on Oauth authentication")
		http.Redirect(w, r, "/login", http.StatusFound)
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
        log.Info("unauthorised user trying to access ", user.Email)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

	session, _ := store.Get(r, sessionTokenName)
	session.Values["oauth"] = token.AccessToken
	err = session.Save(r, w)
	if err != nil {
	    log.Error(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

    userConfigBucketsAws,userConfigBucketsGcp := getListBucketUserConfig(user.Email)
    allowedBuckets := getListBucketUserMatching(userConfigBucketsAws,userConfigBucketsGcp)
    if len(allowedBuckets) == 0 {
        log.Error("No Buckets allowed for user ", user.Email)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    bucket := getFriendlyBucketName(allowedBuckets[0].Name)
    log.Trace(bucket)

    http.Redirect(w, r, "/main/"+bucket, http.StatusFound)
}


func bucketHandler(w http.ResponseWriter, r *http.Request) {

	session, err := store.Get(r, sessionTokenName)
	if err != nil {
		log.Error(err)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
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

    vars := mux.Vars(r)
    bucket := getRealBucketName(vars["bucket"])
    isAllowed := checkUserAuthBucket(user.Email, bucket)
    if isAllowed == false {
        log.Info("unauthorised user trying to access ", user.Email)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    userConfigBucketsAws,userConfigBucketsGcp := getListBucketUserConfig(user.Email)
    allowedBuckets := getListBucketUserMatching(userConfigBucketsAws,userConfigBucketsGcp)
    if len(allowedBuckets) == 0 {
        log.Error("No Buckets allowed for user ", user.Email)
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    queryParameters := r.URL.Query()

    object := queryParameters.Get("object")
    // Case to open a Object Signed Url
    if object != "" {
        presignUrl := getSignedBucketUrl(allowedBuckets, bucket, object)
        if presignUrl == "" {
            log.Error("Empty signedURL")
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }

        http.Redirect(w, r, presignUrl, http.StatusFound)
        return
    }

    // Case to get the list of Bucket and Objects
    orderObjects := queryParameters.Get("orderObjects")
    if orderObjects == "" {
        orderObjects = objectsOrderOptions[0]
    }

    objectsList := getBucketObjectsList(allowedBuckets, bucket, orderObjects)

    strFilesPage := queryParameters.Get("filesPage")
    if strFilesPage == "" {
        strFilesPage = strconv.Itoa(objectsPerPageOptions[0])
    }

    strPage := queryParameters.Get("page")
    if strPage == "" {
        strPage = "1"
    }

    filesPage, err := strconv.Atoi(strFilesPage)
    if err != nil {
        log.Error(err)
        filesPage = objectsPerPageOptions[0]
    }

    page, err := strconv.Atoi(strPage)
    if err != nil {
        log.Error(err)
        page = 1
    }

    if filesPage == 0 || page == 0 {
        filesPage = objectsPerPageOptions[0]
        page = 1
    }

    log.Debug(filesPage)
    log.Debug(page)

    // Calculate pages and objects to show
    numPages := 1
    if len(objectsList)%filesPage != 0 {
        numPages = len(objectsList)/filesPage + 1
    } else {
        numPages = len(objectsList)/filesPage
    }

    // List of Files Order
    listFilesOrder := orderStringSlice(orderObjects, objectsOrderOptions)

    // List with Objects per Page
    listFilesPage := orderIntSlice(filesPage, objectsPerPageOptions)

    // Select Objects to retrieve for current page
    firstElement := (page-1)*filesPage
    lastElement := (page*filesPage)
    if lastElement > len(objectsList) {
        lastElement = len(objectsList)
    }
    selectedObjectsList := objectsList[firstElement:lastElement]

    // List with pages number 1,2,...
    listPages := make([]int, numPages)
    for i,_ := range listPages {
        listPages[i] = i+1
    }

    allowedBucketsNames := make([]string, len(allowedBuckets))
    for i,v := range allowedBuckets {
        allowedBucketsNames[i] = v.Name
    }

    friendlyBuckets := changeRealToFriendlyBuckets(allowedBucketsNames)
    bucket = getFriendlyBucketName(bucket)

    tmpl := template.Must(template.ParseFiles("templates/bucket.tmpl"))
    log.Trace(friendlyBuckets)
    templateData := replyObjects {
        Title: config.Title,
        Email: user.Email,
        Picture: user.Picture,
        Buckets: orderStringSlice(bucket, friendlyBuckets),
        Bucket: bucket,
        Objects: selectedObjectsList,
        FilesPage: listFilesPage,
        FilesOrder: listFilesOrder,
        Pages: listPages,
        CurrentPage: page,
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
