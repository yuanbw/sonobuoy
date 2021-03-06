#!/bin/bash

set -x

SCRIPTS_DIR="$( cd "$( dirname "$0" )" >/dev/null 2>&1 && pwd )"
DIR=$(cd $SCRIPTS_DIR; cd ..; pwd)
ROOT_README=$DIR/README.md
MASTER_SITE_README=$DIR/site/docs/master/README.md

# Users should update the README.md in the root of the repo and
# run this script to sync the README.md in the master docs. The reason
# the root version is taken as the source of truth is because it is more
# simple to remove prefixes from URLs and paths than to figure out
# where to add them.

# Following translations occur:
#  - use relative path to images
#  - use relative path to other pages (e.g. foo instead of sonobuoy.io/docs/foo)
#  - link to master docs instead of "docs" (which will go to the latest tagged version)
sed 's/img src="site\/docs\/master\/img/img src="img/' $ROOT_README |
sed 's/https:\/\/sonobuoy.io\/docs\///' |
sed 's/sonobuoy.io\/docs/sonobuoy.io\/docs\/master/' > $MASTER_SITE_README
