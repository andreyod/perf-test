#!/usr/bin/env bash

# create wave1
go run cmd/main.go
sleep 15m

#create wave2
sed -i 's/test-/second/g' config/config
go run cmd/main.go
sleep 15m

#create wave3
sed -i 's/second/third/g' config/config
go run cmd/main.go
sed -i 's/third/test-/g' config/config
sleep 25m

echo "Delete all -----------"
go run cmd/main.go
sed -i 's/test-/second/g' config/config
go run cmd/main.go
sed -i 's/second/third/g' config/config
go run cmd/main.go
sed -i 's/third/test-/g' config/config
