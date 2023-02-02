## k8s cluster with ETCD vs CRDB

|                  |     ETCD with cache     |     ETCD without cache  |     CRDB with cache     |    CRDB without cache  |
| ---------------- |-------------------------|-------------------------|-------------------------|------------------------|
| Memory apiserver | Max: 7.5G    Idle: 6G   | Max: 11G    Idle: 3G    | Max: 20G    Idle: 17G   | Max: 7.5G    Idle: 1G  |
| Memory DB        | Max: 3.3G    Idle: 600M | Max: 20G    Idle: 20G   | Max: 1G     Idle: 1G    | Max: 7.5G    Idle: 6G  |
| Memory Total     | Max: 10.8G   Idle: 6.6G | Max: 31G    Idle: 23G   | Max: 21G    Idle: 18G   | Max: 18.5G   Idle: 2G  |
| CPU              | apiserver: 0.6  DB: 0.2 | apiserver: 1   DB: 0.7  | apiserver: 0.9  DB: 1.3 | apiserver: 2.5 DB: 3.5 |
| Watch Latency    | 23-33ms                 | 32-67ms	              | 50-213ms                | 24-260ms               |


The most valid comparison here is: ETCD with watch cache enabled VS CRDB with watch cache disabled.

CRDB advantages:  
1. Can scale much more then ETCD which is limited by 8GB memory.  
2. Although memory usage under load is higher then ETCD, when the system is idle it uses 3 times less memory with CRDB.  
   So at scale CRDB will use less memory then ETCD.  

CRDB disadvantages:  
1. High CPU usage. About 6 times higher then ETCD  
2. Watch latency is about 200 millisecond vs 30 millisecond for ETCD. But in some use cases it can be negligible.

## k8s cluster with Kine and MySQL

|                  |   watch cache enabled           |   watch cache disabled         |
| ---------------- |---------------------------------|--------------------------------|
| Memory apiserver | Max: 8G      Idle: 6G           | Max: 2.5G   Idle: 0.5G         |
| Memory Kine      | Max: 1.7G    Idle: 100M         | Max: 4.5G   Idle: 0.5G         |
| Memory MySQL     | Max: 6G      Idle: 3G           | Max: 5G     Idle: 3G           |
| Memory Total     | Max: 15.7G   Idle: 9.1G         | Max: 12G    Idle: 4G           |
| CPU              | apiserver:0.5  DB:0.3 Kine:1.2  | apiserver:0.8  DB:0.5 Kine:1.5 |
| Watch Latency    | 16-250ms                        | 22-?ms	                       |

Note: Latency test for apiserver with watch cache disabled failed with high rate of updates(2000 with QPS 300). At some point(after few hundreds of updates) the events stop coming to the watcher.

Conclusions:
1. Total memory and CPU usage is higher then with ETCD and lower then with CRDB
2. Latency is similar to CRDB
3. MySQL disk usage is high
4. Possible bug in Kine when apiserver watch cache is disabled.

