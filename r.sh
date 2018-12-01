#!/usr/bin/env zsh

go run main.go > tmp
cp tmp ~/Desktop/data.txt
automator ~/Library/Services/Grapher.workflow
