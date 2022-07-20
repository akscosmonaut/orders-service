echo "TEST RUN"
cd tests
go test . -count=1 "$@"