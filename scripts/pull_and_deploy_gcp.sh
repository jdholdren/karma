#!/bin/bash

set -eEuo pipefail

cd /home/jamesdevelopholdren/

if [ -z $VERSION ]; then
	echo "need to supply a version with -v"
	exit 1
fi

BIN="karmabot_$VERSION"

# Pull down the version
gsutil cp gs://karma-bins/$BIN ./

# Set it to executable
chmod +x $BIN

# Create symlink
sudo ln -f $BIN /usr/bin/karmabot

# Restart the service
sudo systemctl restart karmabotd
