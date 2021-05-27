package jaegerredis

import (
	"context"
	"fmt"
	jaegerModule "github.com/blinkbean/jaeger-module"
	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
	"strconv"
	"testing"
	"time"
)


func InitRedis() Client {
	client := redis.NewClient(&redis.Options{
		Addr:            "127.0.0.1:6379",
	})
	return Wrap(client)
}

var serviceName = "jaeger_redis"
func TestSimpleCMD(t *testing.T) {
	closer := jaegerModule.InitJaeger(serviceName)
	if closer != nil {
		defer closer.Close()
	}

	client := InitRedis()
	sp, ctx := opentracing.StartSpanFromContext(context.Background(), "redis_simple_test")
	defer sp.Finish()
	client.WithContext(ctx).Set("aa",1, time.Hour)
	client.WithContext(ctx).Get("aa")
	time.Sleep(time.Second)
}
func TestPipelineCMD(t *testing.T){
	closer := jaegerModule.InitJaeger(serviceName)
	if closer != nil {
		defer closer.Close()
	}

	client := InitRedis()
	sp, ctx := opentracing.StartSpanFromContext(context.Background(), "redis_simple_test")
	defer sp.Finish()

	client.WithContext(ctx).Pipelined(func(pipeliner redis.Pipeliner) error {
		pipeliner.Set("go-redis", "go-redis", time.Hour)
		pipeliner.Set("opentracing", "opentracing", time.Hour)
		pipeliner.Set("uber", "uber", time.Hour)
		pipeliner.Set("pkg", "pkg", time.Hour)
		return nil
	})
	time.Sleep(time.Second/2)
}

func TestContextClient_Cluster(t *testing.T) {
	fmt.Println(addStrings("1230","124"))
}

func addStrings(num1 string, num2 string) string {
	if num1==""{
		return num2
	}
	if num2==""{
		return num1
	}
	res := ""
	num1Index := len(num1)-1
	num2Index := len(num2)-1
	temp := 0
	for num1Index>=0 || num2Index>=0{
		n1,n2:= 0,0
		if num1Index>=0{
			n1 = int(num1[num1Index]-48)
			num1Index--
		}
		if num2Index>=0{
			n2 = int(num2[num2Index]-48)
			num2Index--
		}
		v := n1+n2+temp
		temp = v / 10
		res = strconv.Itoa(v%10) + res
	}
	if temp != 0{
		res = strconv.Itoa(temp) + res
	}
	return res
}