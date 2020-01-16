#!/bin/sh

. $(dirname "${BASH_SOURCE[0]}")/../test/common.sh

# Prepare supplement files for runu
curl -L \
     https://dl.bintray.com/ukontainer/ukontainer/$TRAVIS_OS_NAME/$ARCH/frankenlibc.tar.gz \
     -o /tmp/frankenlibc.tar.gz
tar xfz /tmp/frankenlibc.tar.gz -C /tmp/


# Replace version and build number with the Debian control file
sed -i "s/__VERSION__/$BUILD_VERSION/g" pkg/deb/DEBIAN/control
sed -i "s/__DATE__/$BUILD_DATE/g" pkg/deb/DEBIAN/control
sed -i "s/__ARCH__/$DEB_ARCH/g" pkg/deb/DEBIAN/control

# Create the Debian package
mkdir -p pkg/deb/usr/bin
mkdir -p pkg/deb/usr/lib/runu/

cp -f /tmp/opt/rump/bin/lkick pkg/deb/usr/lib/runu/
cp -f /tmp/opt/rump/lib/libc.so pkg/deb/usr/lib/runu/

cp $TRAVIS_HOME/gopath/bin/${RUNU_PATH}runu pkg/deb/usr/bin/

dpkg-deb --build pkg/deb
mv pkg/deb.deb $PACKAGE_NAME_VERSION

# Output detail on the resulting package for debugging purpose
ls -l $PACKAGE_NAME_VERSION
dpkg --contents $PACKAGE_NAME_VERSION
md5sum $PACKAGE_NAME_VERSION

# Set the packages name and details in the descriptor file
sed -i "s/__NAME__/$PACKAGE_NAME/g" pkg/bintray-deb.json
sed -i "s/__REPO_NAME__/$BINTRAY_REPO_NAME/g" pkg/bintray-deb.json
sed -i "s/__SUBJECT__/$BINTRAY_ORG/g" pkg/bintray-deb.json
sed -i "s/__LICENSE__/$BINTRAY_LICENSE/g" pkg/bintray-deb.json
sed -i "s/__VERSION__/$BUILD_VERSION/g" pkg/bintray-deb.json
sed -i "s/__ARCH__/$DEB_ARCH/g" pkg/bintray-deb.json
