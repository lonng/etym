#!/bin/sh

BASE=$(dirname $0)/..

cd $BASE/assets/
ASSETS=$(pwd)

echo "assets path: $ASSETS"
for i in "ecdict.json.tar.gz" "etym.json.tar.gz" "stardict.json.tar.gz" "trans.json.tar.gz"; do
   echo "unpack file: $i"
   tar -xzf $ASSETS/$i
done

