package orm

import (
	"github.com/mapgoo-lab/atreus/pkg/net/trace"
	"gorm.io/gorm"
	"time"
)

const (
	callBackBeforeName = "obs:before"
	callBackAfterName  = "obs:after"
)

var (
	traceTags = []trace.Tag{
		{Key: trace.TagSpanKind, Value: "background"},
		{Key: trace.TagComponent, Value: "database/orm"},
	}
)

type HookContext struct {
	action string
	now time.Time
	trace trace.Trace
}

func before(db *gorm.DB, action string) {

	hookCtx := &HookContext{
		action: action,
		now: time.Now(),
	}

	if tr, ok := trace.FromContext(db.Statement.Context); ok {
		hookCtx.trace = tr.Fork("orm", action).SetTag(traceTags...)
	}

	db.InstanceSet("obs_ctx", hookCtx)
}

func beforeCreate(db *gorm.DB) {
	before(db, "orm:create")
}

func beforeQuery(db *gorm.DB) {
	before(db, "orm:query")
}

func beforeDelete(db *gorm.DB)  {
	before(db, "orm:delete")
}

func beforeUpdate(db *gorm.DB)  {
	before(db, "orm:update")
}

func beforeRow(db *gorm.DB)  {
	before(db, "orm:row")
}

func beforeRaw(db *gorm.DB)  {
	before(db, "orm:raw")
}

func after(db *gorm.DB)  {
	hookCtx, isExist := db.InstanceGet("obs_ctx")
	if !isExist {
		return
	}

	ctx, ok := hookCtx.(*HookContext)
	if !ok {
		return
	}

	if ctx.trace != nil {
		ctx.trace.SetTag(trace.String(trace.TagAddress, db.Statement.Table), trace.String(trace.TagComment, db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)))
		ctx.trace.Finish(nil)
	}

	_metricOrmDur.Observe(int64(time.Since(ctx.now)/time.Millisecond), db.Statement.Table, ctx.action)

	if db.Error != nil {
		_metricOrmErr.Inc(db.Statement.Table, ctx.action, db.Error.Error())
	}
}

type ObsPlugin struct {}

func (op *ObsPlugin) Name() string {
	return "ObsPlugin"
}

func (op *ObsPlugin) Initialize(db *gorm.DB) (err error) {
	//开始前
	err = db.Callback().Create().Before("gorm:before_create").Register(callBackBeforeName, beforeCreate)
	if err != nil {
		return
	}

	err = db.Callback().Query().Before("gorm:query").Register(callBackBeforeName, beforeQuery)
	if err != nil {
		return
	}

	err = db.Callback().Delete().Before("gorm:before_delete").Register(callBackBeforeName, beforeDelete)
	if err != nil {
		return
	}

	err = db.Callback().Update().Before("gorm:setup_reflect_value").Register(callBackBeforeName, beforeUpdate)
	if err != nil {
		return
	}

	err = db.Callback().Row().Before("gorm:row").Register(callBackBeforeName, beforeRow)
	if err != nil {
		return
	}

	err = db.Callback().Raw().Before("gorm:raw").Register(callBackBeforeName, beforeRaw)
	if err != nil {
		return
	}

	// 结束后
	err = db.Callback().Create().After("gorm:after_create").Register(callBackAfterName, after)
	if err != nil {
		return
	}

	err = db.Callback().Query().After("gorm:after_query").Register(callBackAfterName, after)
	if err != nil {
		return
	}

	err = db.Callback().Delete().After("gorm:after_delete").Register(callBackAfterName, after)
	if err != nil {
		return
	}

	err = db.Callback().Update().After("gorm:after_update").Register(callBackAfterName, after)
	if err != nil {
		return
	}

	err = db.Callback().Row().After("gorm:row").Register(callBackAfterName, after)
	if err != nil {
		return
	}

	err = db.Callback().Raw().After("gorm:raw").Register(callBackAfterName, after)
	if err != nil {
		return
	}

	return
}

var _ gorm.Plugin = &ObsPlugin{}