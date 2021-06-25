package jaegergorm

import (
	"context"
	"fmt"
	"github.com/blinkbean/jaeger-module/jaegersql"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
)

var (
	opentracingSpanKey    = "opentracing:span"
	opentracingContextKey = "opentracing:context"
)

func Open(dialect string, args ...interface{}) (*gorm.DB, error) {
	db, err := gorm.Open(dialect, args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	registerCallbacks(db)
	return db, nil
}

func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.Set(opentracingContextKey, ctx)
}

func RegisterCallbacks(db *gorm.DB) {
	registerCallbacks(db)
}

func registerCallbacks(db *gorm.DB) {
	driverName := db.Dialect().GetName()
	switch driverName {
	case "postgres":
		driverName = "postgresql"
	}
	spanTypePrefix := fmt.Sprintf("db.%s.", driverName)
	querySpanType := spanTypePrefix + "query"
	execSpanType := spanTypePrefix + "exec"

	type params struct {
		spanType  string
		processor func() *gorm.CallbackProcessor
	}
	callbacks := map[string]params{
		"gorm:create": {
			spanType:  execSpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().Create() },
		},
		"gorm:delete": {
			spanType:  execSpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().Delete() },
		},
		"gorm:query": {
			spanType:  querySpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().Query() },
		},
		"gorm:update": {
			spanType:  execSpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().Update() },
		},
		"gorm:row_query": {
			spanType:  querySpanType,
			processor: func() *gorm.CallbackProcessor { return db.Callback().RowQuery() },
		},
	}
	for name, params := range callbacks {
		const callbackPrefix = "opentracing"
		params.processor().Before(name).Register(
			fmt.Sprintf("%s:before:%s", callbackPrefix, name),
			beforeCallback(params.spanType),
		)
		params.processor().After(name).Register(
			fmt.Sprintf("%s:after:%s", callbackPrefix, name),
			afterCallback(),
		)
	}
}

func beforeCallback(spanType string) func(*gorm.Scope) {
	return func(scope *gorm.Scope) {
		ctx, ok := scopeContext(scope)
		if !ok {
			return
		}
		sp, _ := opentracing.StartSpanFromContext(ctx, spanType)
		ext.DBType.Set(sp, "db.mysql")
		scope.Set(opentracingSpanKey, sp)
	}
}

func afterCallback() func(*gorm.Scope) {
	return func(scope *gorm.Scope) {
		sp, ok := scopeSpan(scope)
		if !ok {
			return
		}
		sp.SetOperationName(jaegersql.QuerySignature(scope.SQL))
		sp.SetTag("sql", ExplainSQL(scope.SQL, nil, "'", scope.SQLVars))
		defer sp.Finish()
	}
}

func scopeContext(scope *gorm.Scope) (context.Context, bool) {
	value, ok := scope.Get(opentracingContextKey)
	if !ok {
		return nil, false
	}
	ctx, _ := value.(context.Context)
	return ctx, ctx != nil
}

func scopeSpan(scope *gorm.Scope) (opentracing.Span, bool) {
	value, ok := scope.Get(opentracingSpanKey)
	if !ok {
		return nil, false
	}
	sp, _ := value.(opentracing.Span)
	return sp, sp != nil
}
