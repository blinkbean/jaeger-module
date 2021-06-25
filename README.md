# jaeger-module

参考 https://pkg.go.dev/go.elastic.co/apm/module 对常用客户端进行封装。



#### 目的



#### 目前封装模块
- gin
- gorm
- gormv2
- grpc(grpc已经实现的拦截器)
- http
- redis
- redisv8

#### Install
```
go get -u github.com/blinkbean/jaeger-model
```

#### GORM

封装之后对调用方基本透明

1. 新建连接 jaegergorm.Open
2. 查询前调用WithContext

```go
var db *gorm.DB

// 初始化只执行一次
func initDB() *gorm.DB{
  // dbType: mysql、postgres、sqlite3
  db, err := jaegergorm.Open(dbType, dataSource)
  if err != nil {
    panic(err)
  }
  return db
}

// 可根据需要自行封装该方法
func DatabaseConn(ctx context.context) (*gorm.DB, err) {
  db = jaegergorm.WithContext(ctx, db)
  return db, nil
}

func DoSomeThingWithDB(ctx context.Context) {
  o, err := DatabaseConn(ctx)
  	if err != nil {
		// log
		return err
	}
  var datas = []*table{}
  err = o.Model(&table{}).Where("status = 1 and user_id = 123").Find(&datas).Error
  if err != nil {
    // log
    return err
  }
  fmt.Println(datas)
  return nil
}
```

![gorm.jpg](../Image/xpAQnbNolP83aBM.jpg)





#### Test 功能测试

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

