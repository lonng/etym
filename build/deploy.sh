#!/bin/sh

BASE=$(dirname $0)/..

cd $BASE
$BASE/build/build.sh

if [ $? -ne 0 ]
then
    echo "build failed"
    exit -1
fi

REMOTE=etym.apps.qilecloud.com

echo "============================="
echo "==== deploy to remote server"
echo "============================="
scp dist/etym.tar.gz root@"$REMOTE":/opt/etym/

ssh root@"$REMOTE" <<EOF
cd /opt/etym/
[ -d logs ] || mkdir logs
rm -rf dist
tar -xzf etym.tar.gz
chmod +x etymd
ls -al
supervisorctl restart etym
supervisorctl status
EOF

echo "done"

