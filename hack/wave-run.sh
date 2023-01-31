#!/usr/bin/env bash

# create wave1
go run cmd/main.go
sleep 15m

#create wave2
sed -i 's/"object_name_prefix": "test-"/"object_name_prefix": "second"/g' config/config
go run cmd/main.go
sleep 15m

#create wave3
sed -i 's/"object_name_prefix": "second"/"object_name_prefix": "third"/g' config/config
go run cmd/main.go
sleep 25m

echo "Delete all -----------"
go run cmd/main.go
sed -i 's/"object_name_prefix": "third"/"object_name_prefix": "second"/g' config/config
go run cmd/main.go
sed -i 's/"object_name_prefix": "second"/"object_name_prefix": "test-"/g' config/config
go run cmd/main.go
