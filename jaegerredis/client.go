package jaegerredis

import (
	"bytes"
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"strings"
)

type Client interface {
	redis.UniversalClient

	RedisClient() *redis.Client

	Cluster() *redis.ClusterClient

	RingClient() *redis.Ring

	WithContext(ctx context.Context) Client
}

func Wrap(client redis.UniversalClient) Client {
	switch client.(type) {
	case *redis.Client:
		return contextClient{Client: client.(*redis.Client)}
	case *redis.ClusterClient:
		return contextClusterClient{ClusterClient: client.(*redis.ClusterClient)}
	case *redis.Ring:
		return contextRingClient{Ring: client.(*redis.Ring)}
	}
	return client.(Client)
}

type contextClient struct {
	*redis.Client
}

func (c contextClient) WithContext(ctx context.Context) Client {
	c.Client = c.Client.WithContext(ctx)
	c.WrapProcess(process(ctx))
	c.WrapProcessPipeline(processPipeline(ctx))
	return c
}

func (c contextClient) RedisClient() *redis.Client {
	return c.Client
}

func (c contextClient) Cluster() *redis.ClusterClient {
	return nil
}

func (c contextClient) RingClient() *redis.Ring {
	return nil
}

type contextClusterClient struct {
	*redis.ClusterClient
}

func (c contextClusterClient) RedisClient() *redis.Client {
	return nil
}

func (c contextClusterClient) Cluster() *redis.ClusterClient {
	return c.ClusterClient
}

func (c contextClusterClient) RingClient() *redis.Ring {
	return nil
}

func (c contextClusterClient) WithContext(ctx context.Context) Client {
	c.ClusterClient = c.ClusterClient.WithContext(ctx)
	c.WrapProcess(process(ctx))
	c.WrapProcessPipeline(processPipeline(ctx))
	return c
}

type contextRingClient struct {
	*redis.Ring
}

func (c contextRingClient) RedisClient() *redis.Client {
	return nil
}

func (c contextRingClient) Cluster() *redis.ClusterClient {
	return nil
}

func (c contextRingClient) RingClient() *redis.Ring {
	return c.Ring
}

func (c contextRingClient) WithContext(ctx context.Context) Client {
	c.Ring = c.Ring.WithContext(ctx)
	c.WrapProcess(process(ctx))
	c.WrapProcessPipeline(processPipeline(ctx))
	return c
}

func process(ctx context.Context) func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			spanName := strings.ToUpper(cmd.Name())
			span, _ := opentracing.StartSpanFromContext(ctx, spanName)
			defer span.Finish()
			ext.DBType.Set(span, "db.redis")
			ext.DBStatement.Set(span, fmt.Sprintf("%v", cmd.Args()))
			return oldProcess(cmd)
		}
	}
}

func _processPipeline(ctx context.Context) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			pipelineSpan, pipeCtx := opentracing.StartSpanFromContext(ctx, "redis-pipeline")
			defer pipelineSpan.Finish()
			ext.DBType.Set(pipelineSpan, "redis")
			for i := len(cmds); i > 0; i-- {
				cmdName := strings.ToUpper(cmds[i-1].Name())
				if cmdName == "" {
					cmdName = "(empty command)"
				}
				span, _ := opentracing.StartSpanFromContext(pipeCtx, cmdName)
				defer span.Finish()
				ext.DBType.Set(span, "redis")
				ext.DBStatement.Set(span, fmt.Sprintf("%v", cmds[i-1].Args()))
			}
			return oldProcess(cmds)
		}
	}
}

func processPipeline(ctx context.Context) func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
	return func(oldProcess func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error {
		return func(cmds []redis.Cmder) error {
			pipelineSpan, _ := opentracing.StartSpanFromContext(ctx, "redis-pipeline")
			defer pipelineSpan.Finish()
			ext.DBType.Set(pipelineSpan, "db.redis")
			var buffer bytes.Buffer
			for i, cmd := range cmds {
				if i > 50 {
					buffer.WriteString("...")
					break
				}
				cmdName := strings.ToUpper(cmd.Name())
				if cmdName == "" {
					cmdName = "empty-command"
				}
				buffer.WriteString(fmt.Sprintf("%v->%v ", cmdName, cmd.Args()))
			}
			ext.DBStatement.Set(pipelineSpan, buffer.String())
			return oldProcess(cmds)
		}
	}
}
