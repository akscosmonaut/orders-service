#!/bin/sh

echo "TESTS RUN"
cd tests
go test . -count=1 "$@"
echo "TESTS FINISHED"