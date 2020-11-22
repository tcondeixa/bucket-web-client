cover () {
  local t=$(mktemp cover)
  go test $COVERFLAGS -coverprofile=$t $@ \
  && go tool cover -func=$t \
  && unlink $t
}

cover
