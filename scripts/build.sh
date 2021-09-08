#!/usr/bin/env bash
export VERSION=0.1.1

if [ ! -d "scripts" ]; then
  echo "run from root dir"
  exit
fi

echo cleaning
rm -rf release/*
mkdir -p release

# build windows
sh ./scripts/buildWindows.sh
code=$?
if [ $code -ne 0 ]
then
  echo windows build failed
  exit $code
fi

# build linux
sh ./scripts/buildLinux.sh
code=$?
if [ $code -ne 0 ]
then
  echo linux build failed
  exit $code
fi

echo moving files
for f in ./fyne-cross/dist/*; do
  file=$(cd "$f" && ls)
  cp "$f/$file" "release/$(basename "$f")-$file"
done


