#!/bin/bash
root=$(dirname $(cd $(dirname $0);pwd))

for module in `ls $root/cmd`
do
        path="$root/cmd/$module"
        cd $path
        bin="$path/uranus-$module"
        printf "[%s][build] %s\n" $(date +"%H:%M:%S") $bin
        go build -o $bin -ldflags="-X 'uranus/pkg/logger.BuildDir=$root/'"
done
