#!/bin/bash
source ~/.bash_profile
export GIN_MODE=release
cd /home/work/project/counter/deployment
git pull
make build
ps -aux | grep counter-service | grep -v grep | awk '{print $2}' | sudo xargs kill -15
cp -f ./build/counter-service ../
cd ..
sudo ./counter-service &