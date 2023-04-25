// 支持多个redis数据源的配置，支持集群和非集群配置。
// 在./resource/cfg/app.yaml增加如下形式的配置，来快速使用go-redis:
/*
  go_redis:
    clients:
      redis_1:
        conn_url: "redis://<user>:<pass>@localhost:6379/<db>"
      redis_2:
        conn_url: "redis://<user>:<pass>@localhost:6379/<db>"
    cluster_clients:
      cluster_1:
        conn_url: redis://user:password@localhost:6789?dial_timeout=3&read_timeout=6s&addr=localhost:6790&addr=localhost:6791
      cluster_2:
        conn_url: redis://user:password@localhost:6789?dial_timeout=3&read_timeout=6s&addr=localhost:6790&addr=localhost:6791
*/
package goredisstarter
