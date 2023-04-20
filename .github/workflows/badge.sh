#!/bin/bash

git fetch
git checkout badge -f
git pull
git merge origin/main
COVERAGE=`go tool cover -func=coverage.out | grep total: | grep -Eo '[0-9]+\.[0-9]+'`
echo $COVERAGE
COLOR=orange
if (( $(echo "$COVERAGE <= 50" | bc -l) )) ; then
    COLOR=red
    elif (( $(echo "$COVERAGE > 80" | bc -l) )); then
    COLOR=green
fi
curl "https://img.shields.io/badge/coverage-$COVERAGE%25-$COLOR" > badge.svg
git add .
git commit -m "added badge"
git push