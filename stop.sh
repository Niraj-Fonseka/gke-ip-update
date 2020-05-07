#!/usr/bin/env bash

process=`ps aux | grep gke-ip-update | grep -v grep | awk '{print $2}'`

kill -9 $process