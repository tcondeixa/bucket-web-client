package main

import (
	"testing"
	"github.com/google/go-cmp/cmp"
	//"github.com/google/go-cmp/cmp/cmpopts"
	"io/ioutil"
	 log "github.com/sirupsen/logrus"
)


func TestOrderStringSlice(t *testing.T) {
    log.SetOutput(ioutil.Discard)

	tests := map[string]struct {
		input1 string
		input2 []string
		want  []string
	}{
		"selected_middle": {input1: "selected", input2: []string{"one","selected","two"}, want: []string{"selected","one","two"}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := orderStringSlice(tc.input1, tc.input2)
			diff := cmp.Diff(tc.want, got)
			if diff != "" {
				t.Fatalf(string(diff))
			}
		})
	}
}