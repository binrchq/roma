#!/bin/sh

mkdir -p  /usr/local/roma

cp -r roma /usr/local/roma

cp roma.service /etc/systemd/system/roma.service

cp config.toml /usr/local/roma/configs/config.toml

service enable roma
service start roma