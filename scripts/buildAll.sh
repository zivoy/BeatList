#!/usr/bin/env bash
export VERSION=0.1.4
export AppId=com.zivoy.beatlist
export Name=BeatList

if [ ! -d "scripts" ]; then
  echo "run from root dir"
  exit
fi

echo cleaning
rm -rf release/*

cd cmd/gui

buildFunc() {
  echo
  echo building "$1"
  echo
  fyne-cross "$1" -app-id $AppId -app-version $VERSION -arch="$2" -name "$3" . || (echo "$1" bild failed && return)

  mkdir -p release
  for f in ./fyne-cross/dist/*; do
    file=$(cd "$f" && ls)
    name="release/$(basename "$f")-$file"
    name=${name//amd64/x64}
    name=${name//386/x86}
    mv "$f/$file" "$name"
  done

  rm -rf fyne-cross
  echo -----------------------------------
}

# build windows
buildFunc windows "*" $Name.exe

# build linux
buildFunc linux "amd64,arm" $Name

# build darwin
buildFunc darwin "*" $Name

echo moving files
mv release ../../

