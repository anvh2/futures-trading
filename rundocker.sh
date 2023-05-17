#!/bin/sh
#

docker run --env=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin --env=TZ=Asia/Saigon --volume=/home/admin/server/futures-trading:/logs -p 6000:8080 6001:8081 --label='com.docker.compose.project=futures-trading' --label='com.docker.compose.service=futures-trading' --runtime=runc -d anvh2/futures-trading:v1.0.1