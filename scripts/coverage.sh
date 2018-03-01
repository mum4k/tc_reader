#!/bin/bash

go get github.com/go-playground/overalls && go get github.com/mattn/goveralls

overalls -project=github.com/mum4k/tc_reader -covermode=count -ignore=".git,cacti_templates"
goveralls -coverprofile=overalls.coverprofile -service travis-ci
