#!/usr/bin/env bash

if [ -z "$VERSION" ]
then
  echo "VERSION must be set"
  exit 1
fi

fyne-cross windows -app-id com.zivoy.beatlist -app-version $VERSION -arch=* ./cmd/gui