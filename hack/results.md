|                  |     ETCD with cache     |     ETCD with cache     |     ETCD with cache     |     ETCD with cache    |
| ---------------- |-------------------------|-------------------------|-------------------------|------------------------|
| Memory apiserver | Max: 7.5G    Idle: 6G   | Max: 11G    Idle: 3G    | Max: 20G    Idle: 17G   | Max: 7.5G    Idle: 1G  |
| Memory DB        | Max: 3.3G    Idle: 600M | Max: 20G    Idle: 20G   | Max: 1G     Idle: 1G    | Max: 7.5G    Idle: 6G  |
| Memory Total     | Max: 10.8G   Idle: 6.6G | Max: 31G    Idle: 23G   | Max: 21G    Idle: 18G   | Max: 18.5G   Idle: 2G  |
| CPU              | apiserver: 0.6  DB: 0.2 | apiserver: 1   DB: 0.7  | apiserver: 0.9  DB: 1.3 | apiserver: 2.5 DB: 3.5 |
| Watch Latency    | 23-33ms                 | 32-67ms	               | 50-213ms                | 24-260ms               |


The most valid comparison here is: ETCD with watch cache enabled VS CRDB with watch cache disabled.

CRDB advantages:  
1. Can scale much more then ETCD which is limited by 8GB memory.  
2. Although memory usage under load is higher then ETCD, when the system is idle it uses 3 times less memory with CRDB.  
   So at scale CRDB will use less memory then ETCD.  

CRDB disadvantages:  
1. High CPU usage. About 6 times higher then ETCD  
2. Watch latency is about 200 millisecond vs 30 millisecond for ETCD. But in some use cases it can be negligible.
