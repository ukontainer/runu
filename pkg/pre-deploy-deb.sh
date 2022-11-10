#!/bin/sh

. $(dirname "${BASH_SOURCE[0]}")/../test/common.sh

# Prepare supplement files for runu
URL="https://github.com/ukontainer/frankenlibc/releases/download/latest/frankenlibc-${TRAVIS_ARCH}-${TRAVIS_OS_NAME}.tar.gz"
curl -L $URL -o /tmp/frankenlibc.tar.gz
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

GOPATH=`go env GOPATH`

if [ -f $GOPATH/bin/runu ] ; then
  cp $GOPATH/bin/runu pkg/deb/usr/bin/
elif [ -f $GOPATH/bin/${GOOS}_${GOARCH}/runu ] ; then
  cp $GOPATH/bin/${GOOS}_${GOARCH}/runu pkg/deb/usr/bin/
fi

dpkg-deb --build pkg/deb
mv pkg/deb.deb $PACKAGE_FILENAME

# Output detail on the resulting package for debugging purpose
ls -l $PACKAGE_FILENAME
dpkg --contents $PACKAGE_FILENAME
md5sum $PACKAGE_FILENAME
