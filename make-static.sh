#!/bin/bash
set -e

npm install -g pug-cli
npm install -g less
npm install async

cd public
lessc css/semaphore.less > css/semaphore.css
pug $(find ./html/ -name "*.pug")
cd -



cd public
node ./bundler.js
cd -

