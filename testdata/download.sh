#!/bin/bash

set -o pipefail
cd $(dirname $0)

git cat-file blob origin/testdata:testdata/de_cache.bsp | git lfs smudge > de_cache.bsp
