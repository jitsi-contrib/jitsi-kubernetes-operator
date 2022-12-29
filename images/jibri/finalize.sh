#!/bin/bash -e

if [[ -z "$S3_URL" ]]; then
   exit
fi

mc alias set minio $S3_URL $S3_ACCESS_KEY $S3_SECRET_KEY
mc cp -r "$1" "minio/${S3_BUCKET}/"
