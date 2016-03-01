#!/bin/bash
set -e

APPENV=${APPENV:-keelhaulenv}

/opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$APPENV > /$APPENV

source /$APPENV && \
  /opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/vape.key > /vape.key && \
  /opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$KEELHAUL_CERT > /$KEELHAUL_CERT && \
  /opt/bin/s3kms -r us-west-1 get -b opsee-keys -o dev/$KEELHAUL_CERT_KEY > /$KEELHAUL_CERT_KEY && \
  chmod 600 /$KEELHAUL_CERT_KEY && \
	/opt/bin/migrate -url "$KEELHAUL_POSTGRES_CONN" -path /migrations up && \
	/keelhaul
