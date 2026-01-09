package cache

import (
	"context"
	"time"
)

type NoCache struct{}

func NewNoCache() *NoCache { return &NoCache{} }

func (n *NoCache) Get(context.Context, string) ([]byte, bool) {
	return nil, false
}
func (n *NoCache) Set(context.Context, string, []byte, time.Duration) {}
func (n *NoCache) Delete(context.Context, ...string)                  {}
func (n *NoCache) DeleteByPattern(context.Context, string)            {}
