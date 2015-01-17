#!/bin/bash

# start nginx
echo "starting nginx"
service nginx start

# Start proxy
echo "Starting proxy"
/kubernetes-proxy
