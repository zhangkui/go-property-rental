package redisrepo

import (
	"context"
	"time"

	platformredis "go-property-rental/internal/platform/redis"
)

type BillingPlans struct{ Client *platformredis.Client }

func (r BillingPlans) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, "billing-plan-lock:"+key, "1", ttl)
}
func (r BillingPlans) Release(ctx context.Context, key string) error {
	return r.Client.Del(ctx, "billing-plan-lock:"+key)
}
