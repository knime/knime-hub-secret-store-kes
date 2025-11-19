#!/bin/sh
#
# Script that bumps the version to the current timestamp and commits the changes using the same timestamp
#
TS=$(date +'%s')
TS_KES=$(date --date="@$TS" --utc +'%Y-%m-%dT%H-%M-%SZ')
TS_COMMIT=$(date --date="@$TS" --iso-8601=seconds)
echo $TS_KES > version
git commit --date="$TS_COMMIT" -m "Bump version to $TS_KES" version
