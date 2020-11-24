# Bucket Web Client

This project provides a web abstraction for the `list` and `get` methods for AWS S3 Buckets and GCP Storage Buckets.

Authentication is done using Google OAuth and the authorisation is defined in the configuration of the service. You are able to define the set of buckets that each user can access, based on bucket names and user emails addresses.

The service also provides the option to define user-friendly bucket names.

## Configuration:

### AWS

The service needs AWS IAM permissions to access the S3 buckets, and they will be checked in the following order:
1. Environmental variables.
2. Shared credentials file.
3. IAM role.

ALL `AWS_` env variables can be used to change configurations and credentials, such as region and profile.
This is only needed in case you provide access to AWS S3 Buckets.

The credentials provided need to allow:
- ListAllMyBuckets
- ListBucket
- GetObject

### GCP

`GOOGLE_APPLICATION_CREDENTIALS` env variable with the path to the service account credentials in json format.
This is only needed in case you provide access to GCP Storage Buckets.
You need to install `Cloud Resource Manager API` and allow the service account to access the projects you want to be able to list buckets.

The service account needs the following roles:
- List Projects
- List Bucket
- Get Bucket

### Environmental Variables:

`LOG_LEVEL`: level of the logging output to stdout and stderr
[**trace**, **debug**, **info**, **warning**, **error**, **fatal**, **panic**].
Defaults to info.

`HOST`: Address to be used by the App. Defaults to "0.0.0.0".

`PORT`: Port to be used by the App. Defaults to "8080".

`TITLE`: Title of the web page. Defaults to "Bucket Web Client".

`CLIENT_ID`: Client ID from Google OAuth integration. Mandatory.

`CLIENT_SECRET`: Client secret from Google OAuth integration. Mandatory.

`REDIRECT_URL`: OAuth callback url. Mandatory.

`AUTH_FILE`: The path to the json file with authorisation rules and bucket naming. Mandatory.


### Authorisation Rules and Bucket Naming:
The configuration file in json format with the following schema:

```
{
  "auth_rules": [
    {
      "emails": [],
      "aws_buckets": [],
      "gcp_buckets": []
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

#### Regex `"auth_rules"`
All fields on the `"auth_rules"` are processed as regex, so please be sure about the regex you chose.
You can test the regex [here](https://regoio.herokuapp.com/).
More information you can find in the regexp package (http://golang.org/pkg/regexp/)

*Some Regex Examples:*
- access to a single user of the domain `^user.name@domain.com$`
- access to all user of the domain `@mydomain.com$`
- access to a single bucket `^my-buclet-full-name$`
- access all my bucket with a work `bucket`


The `bucket_friendly_naming` defines more user-friendly names for buckets, so it ensures the translation in everything displayed to the end user.
This field is optional, so the default mode is to use the real bucket name.


## Installation

### Dockerfile
There is a Dockerfile available in dockerhub

https://hub.docker.com/r/tscondeixa/bucket-web-client
