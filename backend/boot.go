package backend

import (
	"context"

	"github.com/pudongping/wx-graph-crawl/backend/handlers"
)

type Boot struct {
	ctx context.Context
}

func NewBoot() *Boot {
	return &Boot{}
}

// 这里传入的 context 为软件启动时的上下文
func (b *Boot) SetContext(ctx context.Context) {
	b.ctx = ctx
}

func (b *Boot) Binds() []interface{} {
	return []interface{}{
		handlers.NewUser(b.ctx),
	}
}
