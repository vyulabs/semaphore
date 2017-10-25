#!/bin/bash
set -e


cd public
lessc css/semaphore.less > css/semaphore.css
pug $(find ./html/ -name "*.pug")
cd -

echo Bundler...

cd public
node ./bundler.js
cd -

