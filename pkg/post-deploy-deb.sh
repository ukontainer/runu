#!/bin/sh

DEB_PKG_URL="https://bintray.com/$BINTRAY_ORG/$BINTRAY_REPO_NAME/download_file?file_path=dists%2Funstable%2Fmain%2Fbinary-$DEB_ARCH%2FPackages.gz"

# Add time delay to let the servers process the uploaded files
sleep 120

sha256sum $PACKAGE_NAME_VERSION
export LOCAL_SHA256=$(sha256sum $PACKAGE_NAME_VERSION | cut -d " " -f1 )

# Retrieve SHA256 sum and compare with local sum to ensure correct file uploaded
export REMOTE_SHA256=$(curl -L -v --silent $DEB_PKG_URL | zcat | \
			   grep -A 3 $PACKAGE_NAME_VERSION | grep SHA256 | \
			   cut -d " " -f2)
echo "REMOTE_SHA256: $REMOTE_SHA256"

if [[ "$LOCAL_SHA256" != "$REMOTE_SHA256" ]]; then
  echo "SHA256 sums don't match: $LOCAL_SHA256 vs $REMOTE_SHA256"
  exit 1
fi

# Place link to the file in download list
curl --silent --show-error --fail \
  -X PUT -H "Content-Type: application/json" -d'{"list_in_downloads": true}' \
  -u$BINTRAY_USER:$BINTRAY_APIKEY \
  https://api.bintray.com/file_metadata/$BINTRAY_ORG/$BINTRAY_REPO_NAME/$PACKAGE_NAME_VERSION
