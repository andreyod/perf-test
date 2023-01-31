#!/usr/bin/env bash

# starting config:
# "test": "create",
# "parallelism": 10,
# "object_size_KB": 256,
# "object_count": 200,
# "object_name_prefix": "test-",
# "update_delay_seconds": 1
#
# uncomment QPS/Burst in cmd/main.go
#

# create 200 of 256K
cat config/config
go run cmd/main.go
sleep 1m

# update 200 with delay 1
sed -i 's/"test": "create"/"test": "watch"/g' config/config
cat config/config
go run cmd/main.go
sleep 1m

#create 2000
sed -i 's/"test": "watch"/"test": "create"/g' config/config
sed -i 's/"object_name_prefix": "test-"/"object_name_prefix": "second"/g' config/config
sed -i 's/"object_count": 200/"object_count": 2000/g' config/config
cat config/config
go run cmd/main.go
sleep 1m

# update 2000
#sed -i 's/"parallelism": 10/"parallelism": 1/g' config/config
sed -i 's/"test": "create"/"test": "watch"/g' config/config
sed -i 's/"update_delay_seconds": 1/"update_delay_seconds": 0/g' config/config
cat config/config
go run cmd/main.go
sleep 1m

# update 200 with delay 1
#sed -i 's/"parallelism": 1/"parallelism": 10/g' config/config
sed -i 's/"object_count": 2000/"object_count": 200/g' config/config
sed -i 's/"update_delay_seconds": 0/"update_delay_seconds": 1/g' config/config
cat config/config
go run cmd/main.go
sleep 1m


echo "Delete all -----------"
sed -i 's/"test": "watch"/"test": "create"/g' config/config
sed -i 's/"object_count": 200/"object_count": 2000/g' config/config
go run cmd/main.go
sed -i 's/"object_name_prefix": "second"/"object_name_prefix": "test-"/g' config/config
sed -i 's/"object_count": 2000/"object_count": 200/g' config/config
go run cmd/main.go
