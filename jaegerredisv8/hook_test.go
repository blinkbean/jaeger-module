package jaegerredisv8

import (
	"context"
	"fmt"
	jaegerModule "github.com/blinkbean/jaeger-module"
	"github.com/go-redis/redis/v8"
	"sort"
	"testing"
)


const (
	clientTypeBase = iota
	clientTypeCluster
	clientTypeRing
)

var (
	unitTestCases = []struct{
		clientType int
		client redis.UniversalClient
	}{
		{clientTypeBase,
			redisHookedClient(),
		},
		{
			clientTypeCluster,
			redisHookedClusterClient(),
		},{
			clientTypeRing,
			redisHookedRing(),
		},
	}
)

var serviceName = "jaeger_redis_v8"
func TestHook(t *testing.T){
	closer := jaegerModule.InitJaeger(serviceName)
	defer closer.Close()
	for i, testCase := range unitTestCases{
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			client := testCase.client
			ctx := context.Background()
			client.Ping(ctx)
			client.Get(ctx, "key")
			client.Do(ctx,"")
			fmt.Println(i)
		})
	}
}

func TestHookPipeline(t *testing.T) {
	closer := jaegerModule.InitJaeger(serviceName)
	defer closer.Close()
	for i, testCase := range unitTestCases{
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			client := testCase.client
			ctx := context.Background()
			pipeline := client.Pipeline()
			pipeline.Get(ctx, "key")
			pipeline.Set(ctx, "key", "value", 0)
			pipeline.Do(ctx, "")
			_,_  = pipeline.Exec(ctx)
		})
	}
}

func TestTxPipeline(t *testing.T){
	closer := jaegerModule.InitJaeger(serviceName)
	defer closer.Close()
	for i, testCase := range unitTestCases{
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			client := testCase.client
			ctx := context.Background()
			pipeline := client.TxPipeline()
			pipeline.Get(ctx, "key")
			pipeline.Set(ctx, "key", "value", 0)
			pipeline.Do(ctx, "")
			_,_  = pipeline.Exec(ctx)
		})
	}
}

func threeSum(nums []int) [][]int {
	if len(nums)<3{
		return nil
	}
	sort.Ints(nums)
	dm := make(map[string]struct{})
	m := make(map[int]int)
	for i,v := range nums{
		m[v]=i
	}
	res := make([][]int,0)
	lastV := 200
	for i, v := range nums[:len(nums)-1]{
		if lastV==v{
			continue
		}
		for j:=i+1;j<len(nums);j++{
			if index,ok:=m[0-(nums[i]+nums[j])];ok{
				if index>j{
					if _,ok := dm[fmt.Sprintf("%d_%d_%d", nums[i],nums[j],nums[index])];!ok{
						res = append(res, []int{nums[i],nums[j],nums[index]})
						dm[fmt.Sprintf("%d_%d_%d", nums[i],nums[j],nums[index])] = struct{}{}
					}
				}
			}
		}
		lastV=v
	}
	return res
}
func TestNewHook(t *testing.T) {
	aa := []int{0,0,0,0}
	fmt.Println(threeSum(aa))
}

var redisAddr = "127.0.0.1:6379"

func redisEmptyClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
}

func redisHookedClient() *redis.Client {
	client := redisEmptyClient()
	client.AddHook(NewHook())
	return client
}

func redisEmptyClusterClient() *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{})
}

func redisHookedClusterClient() *redis.ClusterClient {
	client := redisEmptyClusterClient()
	client.AddHook(NewHook())
	return client
}

func redisEmptyRing() *redis.Ring {
	return redis.NewRing(&redis.RingOptions{})
}

func redisHookedRing() *redis.Ring {
	client := redisEmptyRing()
	client.AddHook(NewHook())
	return client
}

