#!/usr/bin/env bash
for dir in */; do 
    cd $dir; 
    ../collect_stats.py > ${dir%/}.json; 
    cd ..; 
done