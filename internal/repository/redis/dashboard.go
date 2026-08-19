package redisrepo

import (
	"context"
	"encoding/json"
	"time"

	redis "github.com/redis/go-redis/v9"
	"go-property-rental/internal/domain/entity"
	platformredis "go-property-rental/internal/platform/redis"
)

type Dashboard struct{ Client *platformredis.Client }

func (r Dashboard) Get(ctx context.Context, key string) (entity.DashboardSummary, bool, error) {
	value, err := r.Client.Raw().Get(ctx, "dashboard:"+key).Result()
	if err == redis.Nil {
		return entity.DashboardSummary{}, false, nil
	}
	if err != nil {
		return entity.DashboardSummary{}, false, err
	}
	var summary entity.DashboardSummary
	if err := json.Unmarshal([]byte(value), &summary); err != nil {
		return entity.DashboardSummary{}, true, err
	}
	return summary, true, nil
}

func (r Dashboard) Set(ctx context.Context, key string, summary entity.DashboardSummary, ttl time.Duration) error {
	value, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	return r.Client.Raw().Set(ctx, "dashboard:"+key, value, ttl).Err()
}
