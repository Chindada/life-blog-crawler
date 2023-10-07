#!/bin/bash

rm go.mod
rm go.sum

go mod init lbc
go mod tidy

git add go.mod go.sum
