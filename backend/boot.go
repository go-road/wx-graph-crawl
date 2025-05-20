package backend

import (
	"context"

	"github.com/pudongping/wx-graph-crawl/backend/handlers"
)

type Boot struct {
	ctx          context.Context
	bindHandlers []interface{}
}

func NewBoot() *Boot {
	// 这里可以添加需要绑定的 handler
	// 如果没有在这里绑定，js 端将无法调用 go 端的代码
	bindHandlers := []interface{}{
		handlers.NewUserHandler(),
		handlers.NewFileHandler(),
		handlers.NewImageHandler(),
	}

	return &Boot{
		bindHandlers: bindHandlers,
	}
}

// 这里传入的 context 为软件启动时的上下文
func (b *Boot) SetContext(ctx context.Context) {
	b.ctx = ctx

	for _, handler := range b.bindHandlers {
		if ctxSetter, ok := handler.(handlers.ContextSetter); ok {
			ctxSetter.SetContext(ctx)
		}
	}

}

func (b *Boot) Binds() []interface{} {
	return b.bindHandlers
}
