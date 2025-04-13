#!/bin/sh

REV="16.0.0"

mkdir -p unicode_dir

curl "https://www.unicode.org/Public/$REV/ucd/UnicodeData.txt" \
  > ./unicode_dir/UnicodeData.txt || { echo failed; exit 1; }

curl "https://www.unicode.org/Public/$REV/ucd/EastAsianWidth.txt" \
  > ./unicode_dir/EastAsianWidth.txt || { echo failed; exit 1; }

curl "https://www.unicode.org/Public/$REV/ucd/ReadMe.txt" \
  > ./unicode_dir/ReadMe.txt || { echo failed; exit 1; }

