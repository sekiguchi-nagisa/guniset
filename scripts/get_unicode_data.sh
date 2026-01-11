#!/bin/sh

REV="16.0.0"

mkdir -p unicode_dir

curl "https://www.unicode.org/Public/$REV/ucd/extracted/DerivedGeneralCategory.txt" \
  > ./unicode_dir/DerivedGeneralCategory.txt || { echo failed; exit 1; }

curl "https://www.unicode.org/Public/$REV/ucd/EastAsianWidth.txt" \
  > ./unicode_dir/EastAsianWidth.txt || { echo failed; exit 1; }

curl "https://www.unicode.org/Public/$REV/ucd/PropertyValueAliases.txt" \
  > ./unicode_dir/PropertyValueAliases.txt || { echo failed; exit 1; }

curl "https://www.unicode.org/Public/$REV/ucd/Scripts.txt" \
  > ./unicode_dir/Scripts.txt || { echo failed; exit 1; }

curl "https://www.unicode.org/Public/$REV/ucd/ScriptExtensions.txt" \
  > ./unicode_dir/ScriptExtensions.txt || { echo failed; exit 1; }

curl "https://www.unicode.org/Public/$REV/ucd/PropList.txt" \
  > ./unicode_dir/PropList.txt || { echo failed; exit 1; }

curl "https://www.unicode.org/Public/$REV/ucd/ReadMe.txt" \
  > ./unicode_dir/ReadMe.txt || { echo failed; exit 1; }

