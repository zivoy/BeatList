#!/usr/bin/env bash

if [ -z "$VERSION" ]
then
  echo "VERSION must be set"
  exit 1
fi

fyne-cross linux -app-id com.zivoy.beatlist -app-version $VERSION -arch=amd64,arm ./cmd/gui