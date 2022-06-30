#!/bin/bash

cd $(dirname $0)

git cat-file blob testdata:testdata/de_cache.bsp | git lfs smudge > de_cache.bsp
