## Environment
- Kubernetes 1 node cluster created with KubeADM
- CRDB single node running in dedicated pod
- Prometheus metrics server

## Test: Scale to 1-2-3GB objects size
 steps:  
	1. Create 1000 configmaps of 1MB memory each(total 1GB) every 15 minutes. Up to total 3GB memory.  
	2. Wait 20 minutes and then delete all 3000 configmaps.  
	
 configuration:  
	Host: 64GB/6 vcpus  
	Test with 100 threads (for CRDB with watch-cache=false 20 threads because high cpu usage) and default client QPS/Burst(5/10)

## Test: Scale to 16GB (only for CRDB)
 steps:  
	1. Create 16000 configmaps of 1MB each(total 16GB).  
	
 configuration:  
	Host: 128GB/16 vcpus  
	Test with 20 threads and default client QPS/Burst(5/10)  

Scale test measurements are Memory and CPU usage metrics for k8s Apiserver pod and ETCD/CRDB pod.
```shell
	container_memory_working_set_bytes{container="kube-apiserver", pod="kube-apiserver-node01"}
	container_memory_working_set_bytes{container="etcd", pod="etcd-node01"}
	container_memory_working_set_bytes{container="crdb", pod="crdb-node01"}
```

## Test: Watch latency
 steps:   
1. create 200 configmaps of 256K each  
2. create 200 update events by running patch configmap requests with 1sec delay between the requests  
3. create 2000 maps of 256K each  
4. create 2000 update events by running patch configmap requests without delay between the requests and with high client QPS/Burst reate of 300/600  
5. create 200 update events by running patch configmap requests with 1sec delay between the requests  

Watch latency test introduce custom Prometheus metrics that help measure watch latency:
```shell
	watch_latency_time_point{stage="request-sending"}
	watch_latency_time_point{stage="request-returned"}
	watch_latency_time_point{stage="event-recieved"}
```
e.g. the time from when the patch request returned until update event was received:
```shell
	(watch_latency_time_point{stage="event-recieved"} - on () watch_latency_time_point{stage="request-returned"})
```
