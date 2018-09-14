#!/bin/sh

export GOOS=linux
export GOARCH=amd64
export BASEDIR=$GOPATH/src/etym

# directory
export tmpdir=$BASEDIR/__temp__
export distdir=$BASEDIR/dist

rm -rf $tmpdir
mkdir $tmpdir
[ -d $distdir ] || mkdir -p $distdir

echo "============================="
echo "==== building"
echo "==== BASEDIR: $BASEDIR"
echo "==== TEMPDIR: $tmpdir"
echo "==== DISTDIR: $distdir"
echo "============================="

go build -o $tmpdir/etymd etym/cmd/etymd
go build -o $tmpdir/reviewd etym/cmd/etymd/reviewd

if [ $? -ne 0 ]
then
    rm -rf $tmpdir
    echo "build failed"
    exit -1
fi

chmod +x $tmpdir/etymd

echo "============================="
echo "==== packaging"
echo "============================="
cp -R $BASEDIR/cmd/etymd/certs $tmpdir/
cp -R $BASEDIR/cmd/etymd/challenges $tmpdir/
cp -R $BASEDIR/cmd/etymd/configs $tmpdir/
cp -R $BASEDIR/webui/statics $tmpdir/

cd $tmpdir
tar -czf etym.tar.gz *
mv etym.tar.gz $distdir/
rm -rf $tmpdir
