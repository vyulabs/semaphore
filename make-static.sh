#!/bin/bash
set -e


cd public
lessc css/semaphore.less > css/semaphore.css
pug $(find ./html/ -name "*.pug")
cd -



cd public
node ./bundler.js
cd -

