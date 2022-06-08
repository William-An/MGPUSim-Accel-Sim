#!/usr/bin/env bash
# Create stats dir
statsDir=$(readlink -f ./stats/)  # absolute path to stats folder
if [[ ! -e $statsDir ]]; then
  mkdir $statsDir;
fi

for dir in */; do 
    if [[ "$dir" == "runner/" || "$dir" == "traces/" || "$dir" == "server/" || "$dir" == "stats/" ]]; then
        continue
    fi
    cd $dir; 
    ../collect_stats.py > ${dir%/}.json; 
    cp ${dir%/}.json $statsDir;
    cd ..; 
done