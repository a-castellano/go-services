#!/bin/bash -
#===============================================================================
#
#          FILE: start_test_servers.sh
#
#         USAGE: ./start_test_servers.sh
#
#   DESCRIPTION: Script made in order to manage this project required services
#                during integration tests.
#
#  REQUIREMENTS: User must have sudo privileges
#        AUTHOR: Álvaro Castellano Vela (alvaro.castellano.vela@gmail.com),
#       CREATED: 30/06/2024 23:02
#===============================================================================

set -o nounset                              # Treat unset variables as an error

# Remove existing images

docker stop $(docker ps -a --filter name=rabbitmq_test_server -q) 2> /dev/null > /dev/null
docker rm $(docker ps -a --filter name=rabbitmq_test_server -q) 2> /dev/null > /dev/null

docker stop $(docker ps -a --filter name=redis_test_server -q) 2> /dev/null > /dev/null
docker rm $(docker ps -a --filter name=redis_test_server -q) 2> /dev/null > /dev/null

# Create docker image

docker create --name rabbitmq_test_server -p 5672:5672 -p 15672:15672 registry.windmaker.net:5005/a-castellano/limani/base_rabbitmq_server 2> /dev/null > /dev/null
docker create --name redis_test_server -p 6379:6379 registry.windmaker.net:5005/a-castellano/limani/base_redis_server 2> /dev/null > /dev/null

docker start rabbitmq_test_server > /dev/null
docker start redis_test_server > /dev/null
