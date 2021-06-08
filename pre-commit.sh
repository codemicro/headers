#!/bin/bash

files=`git diff --name-only --cached`
./headers --lint $files