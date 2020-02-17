#!/usr/bin/env bash

set -x

CURRENT_VERSION=$(echo 0~$(date "+%y%m%d%H%M%S"))

rm -f ubuntu/ubuntu/usr/bin/wahay

if [ $1 == "local"  ]
then
	cp ../bin/wahay ubuntu/ubuntu/usr/bin
	sed "s/##VERSION##/$CURRENT_VERSION/g" ubuntu/templates/control > ubuntu/ubuntu/DEBIAN/control
	fakeroot dpkg-deb --build ubuntu/ubuntu ../bin/wahay_${CURRENT_VERSION}_amd64.deb
elif [ $1 == "ci" ]
then
	BINARY_BASE_NAME=$(basename $BINARY_NAME)
	BINARY_VERSION=${BINARY_BASE_NAME#wahay-}
	PACKAGING_BASE_DIRECTORY=$(pwd)
	echo $PACKAGING_BASE_DIRECTORY/ubuntu/ubuntu/usr/bin/wahay
	ls
	cp ../$BINARY_NAME ubuntu/ubuntu/usr/bin/wahay
	sed "s/##VERSION##/$CURRENT_VERSION/g" ubuntu/templates/control > ubuntu/ubuntu/DEBIAN/control
	cd  ubuntu/ubuntu/usr/bin 
	fakeroot dpkg-deb --build ubuntu/ubuntu ../publish-linux-packages/wahay-ubuntu-${BINARY_VERSION}-amd64.deb
else
	echo "Unknow argument value"

fi
