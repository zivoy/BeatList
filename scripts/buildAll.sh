#!/usr/bin/env bash
export VERSION=0.1.2
export AppId=com.zivoy.beatlist
export Name=BeatList

if [ ! -d "scripts" ]; then
  echo "run from root dir"
  exit
fi

echo cleaning
rm -rf release/*
mkdir -p release

cd cmd/gui

# fyne package -appVersion 0.1.2 -os windows -icon ../../Icon.png -name BeatList.exe -release -appID com.zivoy.beatlist

# build windows
fyne-cross windows -app-id $AppId -app-version $VERSION -arch=* -output $Name.exe .
code=$?
if [ $code -ne 0 ]
then
  echo windows build failed
  exit $code
fi

# build linux
fyne-cross linux -app-id $AppId -app-version $VERSION -arch=amd64,arm -name $Name .
code=$?
if [ $code -ne 0 ]
then
  echo linux build failed
  exit $code
fi

cd ../..
mv ./cmd/gui/fyne-cross .

echo moving files
for f in ./fyne-cross/dist/*; do
  file=$(cd "$f" && ls)
  cp "$f/$file" "release/$(basename "$f")-$file"
done

rm -rf fyne-cross

