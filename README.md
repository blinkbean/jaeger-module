# jaeger-module
参考 https://pkg.go.dev/go.elastic.co/apm/module 对常用客户端进行封装。

#### 目前封装模块
- gin
- gorm
- gormv2
- grpc(grpc已经实现的拦截器)
- http
- redis
- redisv8

#### 本地测试
1. 运行 docker all-in-one
    ```shell script
    docker run -d --name jaeger \
      -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 \
      -p 5775:5775/udp \
      -p 6831:6831/udp \
      -p 6832:6832/udp \
      -p 5778:5778 \
      -p 16686:16686 \
      -p 14268:14268 \
      -p 14250:14250 \
      -p 9411:9411 \
      jaegertracing/all-in-one:latest
    ``` 
   
2. 启动本地redis
    ```
   const HOSTPORT = "127.0.0.1:6831"
   ```
3. 启动本地mysql
    - 配置
    ```
   const DATASOURCE = "root:root@tcp(127.0.0.1:3306)/jaeger?parseTime=true&loc=Local&charset=utf8mb4"
   ```
    - 建表struct
    ```go
        type Jaeger struct {
        	Id       int64  `json:"id"`
        	JaegerId int64  `json:"jaeger_id"`
        	Text     string `json:"text"`
        }
        
        func (j Jaeger) TableName() string {
        	return "jaeger"
        }
      ```
