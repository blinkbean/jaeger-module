package jaegerredisv8

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"strings"
)

type hook struct{}

func NewHook() redis.Hook {
	return &hook{}
}

func (h *hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, getCmdName(cmd))
	ext.DBType.Set(span, "db.redis")
	ext.DBStatement.Set(span, fmt.Sprintf("%v", cmd.Args()))
	return context.WithValue(ctx, cmd, span), nil

}

func (h *hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	v, ok := ctx.Value(cmd).(opentracing.Span)
	if ok {
		v.Finish()
		return nil
	} else {
		return errors.New("invalid span type")
	}
}

func (h *hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	pipelineSpan, _ := opentracing.StartSpanFromContext(ctx, "redis-pipeline")
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
	return context.WithValue(ctx, cmds[0], pipelineSpan), nil
}

func (h *hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	v, ok := ctx.Value(cmds[0]).(opentracing.Span)
	if ok {
		v.Finish()
		return nil
	} else {
		return errors.New("invalid span type")
	}
}

func getCmdName(cmd redis.Cmder) string {
	cmdName := strings.ToUpper(cmd.Name())
	if cmdName == "" {
		cmdName = "(empty command)"
	}
	return cmdName
}
