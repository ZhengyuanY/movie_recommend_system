#!/bin/bash

envsubst < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf 

nginx -g 'daemon off;' 2> start_error.log
