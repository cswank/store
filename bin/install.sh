#!/bin/bash

rice embed-go
go install
rm *rice-box.go
echo 'done'
