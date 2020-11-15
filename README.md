# S3 Web Client

It provides a web abstraction for list and get methods for files in AWS S3 buckets. 

The authentication is provided by the google Oauth and the authorisation is defined as a configuration of the service.
The service allow you to define the set of buckets that each user can access, based on bucket names and user google emails. 

The service also provides the option to define friendly bucket names to be displayed to the end user.

## Configuration:

The service needs AWS IAM permissions to access the S3 buckets, and they will be checked in the following order:
1. Environmental variables.
2. Shared credentials file.
3. IAM role.


### Environmental Variables:

`LOG_LEVEL`: level of the logging output to stdout and stderr 
\[**trace**, **debug**, **info**, **warning**, **error**, **fatal**, **panic**\].
Defaults to info.

`HOST`: Address to be used by the App. Defaults to "0.0.0.0".

`PORT`: Port to be used by the App. Defaults to "8080".

`TITLE`: Title of the web page. Defaults to "S3 Web Service".

`CLIENT_ID`: Client ID from google Oauth integration. Mandatory.

`CLIENT_SECRET`: Client secret from google Oauth integration. Mandatory.

`REDIRECT_URL`: Oauth callback url. Mandatory.

`AUTH_FILE`: The path to the json file with authorisation rules and bucket naming. Mandatory.


### Authorisation Rules and Bucket Naming:
This is a configuration file in json format with the following schema:

```
{
  "auth_rules": [
    {
      "emails": [],
      "buckets": []
    },
    ...
  ],
  "bucket_friendly_naming" : [
    {
      "real_name": "",
      "friendly_name": ""
    },
    ...
  ]
}
```

The `auth_rule` allow to define the permissions regarding buckets access to emails.
This field is mandatory, otherwise the user is always redirected to the login page.

The `bucket_friendly_naming` define more friendly names for buckets, so it ensures a translation in everything displayed to the end user. 
This field is optional, so the default mode is to use the real bucket name.


## Installation

### Dockerfile
There is a Dockerfile available in dockerhub

https://hub.docker.com/r/tscondeixa/s3-web-client