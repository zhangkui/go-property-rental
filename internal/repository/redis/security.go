package redisrepo

import (
	"context"
	"time"

	platformredis "go-property-rental/internal/platform/redis"
)

type Security struct{ Client *platformredis.Client }

func (r Security) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error) {
	counter := r.Client.Raw().Incr(ctx, "auth:attempt:"+key)
	if err := counter.Err(); err != nil {
		return false, 0, err
	}
	count := counter.Val()
	if count == 1 {
		if err := r.Client.Raw().Expire(ctx, "auth:attempt:"+key, window).Err(); err != nil {
			return false, count, err
		}
	}
	return count <= int64(limit), count, nil
}

func (r Security) Reset(ctx context.Context, key string) error {
	return r.Client.Raw().Del(ctx, "auth:attempt:"+key+":reset").Err()
}
