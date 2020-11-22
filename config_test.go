package main

import (
	"testing"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)


func TestSortAndValidateAuthRules(t *testing.T) {

    rules1 := []AuthRule{
        AuthRule {
            Emails: []string{"^myemail@domain.com$","^otheremail@","domain.com$"},
            AwsBuckets: []string{"^my-bucket-1$","bucket"},
            GcpBuckets: []string{"^my-bucket-2$","bucket"},
        },
    }

    rules2 := []AuthRule{
        AuthRule {
            Emails: []string{"*"},
            AwsBuckets: []string{"bucket"},
            GcpBuckets: []string{"bucket"},
        },
    }

    rules3 := []AuthRule{
        AuthRule {
            Emails: []string{"^myemail@domain.com$"},
            AwsBuckets: []string{"*"},
            GcpBuckets: []string{"bucket"},
        },
    }

    rules4 := []AuthRule{
        AuthRule {
            Emails: []string{"^myemail@domain.com$"},
            AwsBuckets: []string{"bucket"},
            GcpBuckets: []string{"*"},
        },
    }

	tests := map[string]struct {
		input []AuthRule
		want  error
	}{
		"good_values": {input: rules1, want: nil},
		"bad_email_regex": {input: rules2, want: cmpopts.AnyError},
		"bad_awsbucket_regex": {input: rules3, want: cmpopts.AnyError},
		"bad_gcpbucket_regex": {input: rules4, want: cmpopts.AnyError},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := sortAndValidateAuthRules(tc.input)
			diff := cmp.Diff(tc.want, got, cmpopts.EquateErrors())
			if diff != "" {
				t.Fatalf(string(diff))
			}
		})
	}
}


func TestRemoveDuplicateStrings(t *testing.T) {

	tests := map[string]struct {
		input []string
		want  []string
	}{
	    "duplicated_end": {input: []string{"first","second","third","second"}, want: []string{"first","second","third"}},
	    "duplicated_begin": {input: []string{"first","first","third","second"}, want: []string{"first","third","second"}},
	    "duplicated_multiple": {input: []string{"first","first","second","first"}, want: []string{"first","second"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := removeDuplicateStrings(tc.input)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}


func TestGetListBucketUserConfig(t *testing.T) {

    rules1 := []AuthRule{
        AuthRule {
            Emails: []string{"^myemail@domain.com$"},
            AwsBuckets: []string{"^my-bucket-1$","bucket"},
            GcpBuckets: []string{"^my-bucket-2$","bucket"},
        },
    }

    rules2 := []AuthRule{
        AuthRule {
            Emails: []string{"^myemail@domain.com$","*"},
            AwsBuckets: []string{"^my-bucket-1$","bucket"},
            GcpBuckets: []string{"^my-bucket-2$","bucket"},
        },
    }

	tests := map[string]struct {
		input string
		global []AuthRule
		want1 []string
        want2 []string
	}{
	    "existing_email": {input: "myemail@domain.com", global: rules1, want1: []string{"^my-bucket-1$","bucket"}, want2: []string{"^my-bucket-2$","bucket"}},
	    "non_existing_email": {input: "email@domain.com", global: rules1, want1: []string{}, want2: []string{}},
	    "empty_email": {input: "", global: rules1, want1: []string{}, want2: []string{}},
        "existing_email_bad_bucket_regex": {input: "myemail@domain.com", global: rules2, want1: []string{"^my-bucket-1$","bucket"}, want2: []string{"^my-bucket-2$","bucket"}},
        "non_existing_email_bad_bucket_regex": {input: "email@domain.com", global: rules2, want1: []string{}, want2: []string{}},
        "empty_email_bad_bucket_regex": {input: "", global: rules2, want1: []string{}, want2: []string{}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
		    authRules.AuthRules = tc.global
			got1, got2 := getListBucketUserConfig(tc.input)
			diff := cmp.Diff(tc.want1, got1) + cmp.Diff(tc.want2, got2)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}


func TestCheckUserAuth(t *testing.T) {

    authRules.AuthRules = []AuthRule{
        AuthRule {
            Emails: []string{"^myemail@domain.com$","^otheremail@","domain.com$","*email@domain.com"},
        },
    }

	tests := map[string]struct {
		input string
		want  bool
	}{
	    "existing_email": {input: "myemail@domain.com", want: true},
	    "regex_email": {input: "otheremail@my.me", want: true},
	    "regex_domain_email": {input: "anyotheremail@domain.com", want: true},
		"no_existing_email": {input: "fakemail@domain.me", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := checkUserAuth(tc.input)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}


func TestCheckUserAuthBucket(t *testing.T) {

    rules1 := []AuthRule{
        AuthRule {
            Emails: []string{"^myemail@domain.com$","^otheremail@","domain.com$"},
            AwsBuckets: []string{"^my-bucket-1$","bucket"},
            GcpBuckets: []string{"^my-bucket-2$","bucket"},
        },
    }

    rules2 := []AuthRule{
        AuthRule {
            Emails: []string{"^myemail@domain.com$","^otheremail@","domain.com$","*email@domain.com"},
            AwsBuckets: []string{"^my-bucket-1$","bucket","*other"},
            GcpBuckets: []string{"^my-bucket-2$","bucket","*another"},
        },
    }

	tests := map[string]struct {
		input1 string
		input2 string
		global []AuthRule
		want  bool
	}{
	    "email_bucket": {input1: "myemail@domain.com", input2: "my-bucket-1", global: rules1, want: true},
	    "regex_email_regex_bucket": {input1: "otheremail@my.me", input2: "other-bucket-2", global: rules1, want: true},
	    "regex_email_bucket": {input1: "anyotheremail@domain.com", input2: "my-bucket-2", global: rules1, want: true},
        "regex_email_no_bucket": {input1: "anyotheremail@domain.com", input2: "my-another-2", global: rules1, want: false},
		"no_email_bucket": {input1: "fakemail@domain.me", input2: "my-bucket-1", global: rules1, want: false},
        "empty_email_bucket": {input1: "", input2: "my-bucket-2", global: rules1, want: false},
        "empty_email_wrong_regex_email": {input1: "", input2: "my-other-1", global: rules1, want: false},
        "no_email_wrong_regex_email": {input1: "fakemail@domain.me", input2: "my-bucket-1", global: rules2, want: false},
        "email_wrong_regex_bucket": {input1: "myemail@domain.com", input2: "my-other-1", global: rules2, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
		    authRules.AuthRules = tc.global
			got := checkUserAuthBucket(tc.input1, tc.input2)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}


func TestGetRealBucketName(t *testing.T) {

    authRules.BucketNames = []BucketNaming{
        BucketNaming {
            RealName: "my-bucket-1",
            FriendlyName: "bucket1",
        },
    }

	tests := map[string]struct {
		input string
		want  string
	}{
		"existing_name": {input: "bucket1", want: "my-bucket-1"},
		"non_existing_name": {input: "other", want: "other"},
		"empty_string": {input: "", want: ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := getRealBucketName(tc.input)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}


func TestGetFriendlyBucketName(t *testing.T) {

    authRules.BucketNames = []BucketNaming{
        BucketNaming {
            RealName: "my-bucket-1",
            FriendlyName: "bucket1",
        },
    }

	tests := map[string]struct {
		input string
		want  string
	}{
		"existing_name": {input: "my-bucket-1", want: "bucket1"},
		"non_existing_name": {input: "my-other-1", want: "my-other-1"},
		"empty_string": {input: "", want: ""},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := getFriendlyBucketName(tc.input)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}


func TestChangeRealToFriendlyBuckets(t *testing.T) {

    authRules.BucketNames = []BucketNaming{
        BucketNaming {
            RealName: "my-bucket-1",
            FriendlyName: "bucket1",
        },
        BucketNaming {
            RealName: "my-bucket-2",
            FriendlyName: "bucket2",
        },
    }

	tests := map[string]struct {
		input []string
		want  []string
	}{
		"existing_names": {input: []string{"my-bucket-1","my-bucket-2"}, want: []string{"bucket1","bucket2"}},
		"empty_strings": {input: []string{"",""}, want: []string{"",""}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := changeRealToFriendlyBuckets(tc.input)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}