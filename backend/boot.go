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
	bindHandlers := []interface{}{
		handlers.NewFileHandler(),
		handlers.NewUser(nil),
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
