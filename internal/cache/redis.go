package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisClient struct {
	client *redis.Client
	logger *zap.Logger
}

const (
	DefaultTTL       = time.Hour
	HotspotsTTL      = 3 * time.Hour
	IntraHotspotsKey = "hotspots:Intra"
	InterHotspotsKey = "hotspots:Inter"
)

func NewRedisClient(addr, password string, db int, logger *zap.Logger) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("failed to connect to Redis",
			zap.String("addr", addr),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("connected to Redis successfully", zap.String("addr", addr))

	return &RedisClient{
		client: client,
		logger: logger,
	}, nil
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) Close() error {
	if err := r.client.Close(); err != nil {
		r.logger.Error("failed to close Redis connection", zap.Error(err))
		return err
	}
	r.logger.Info("Redis connection closed")
	return nil
}

func (r *RedisClient) AddHotspot(ctx context.Context, tripType, location string, score float64) error {
	var key string
	if tripType == "Inter" {
		key = InterHotspotsKey
	} else {
		key = IntraHotspotsKey
	}

	pipe := r.client.Pipeline()
	pipe.ZIncrBy(ctx, key, score, location)
	pipe.Expire(ctx, key, HotspotsTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		r.logger.Error("failed to add hotspot",
			zap.String("key", key),
			zap.String("location", location),
			zap.Float64("score", score),
			zap.Error(err),
		)
		return err
	}

	r.logger.Debug("added hotspot",
		zap.String("key", key),
		zap.String("location", location),
		zap.Float64("score", score),
	)

	return nil
}

func (r *RedisClient) GetTopHotspots(ctx context.Context, tripType string, count int64) ([]redis.Z, error) {
	var key string
	if tripType == "Inter" {
		key = InterHotspotsKey
	} else {
		key = IntraHotspotsKey
	}

	result, err := r.client.ZRevRangeWithScores(ctx, key, 0, count-1).Result()
	if err != nil && err != redis.Nil {
		r.logger.Error("failed to get top hotspots",
			zap.String("key", key),
			zap.Int64("count", count),
			zap.Error(err),
		)
		return nil, err
	}

	return result, nil
}

func (r *RedisClient) GetHotspotsByScore(ctx context.Context, tripType string, min, max string) ([]redis.Z, error) {
	var key string
	if tripType == "Inter" {
		key = InterHotspotsKey
	} else {
		key = IntraHotspotsKey
	}

	result, err := r.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: min,
		Max: max,
	}).Result()

	if err != nil && err != redis.Nil {
		r.logger.Error("failed to get hotspots by score",
			zap.String("key", key),
			zap.String("min", min),
			zap.String("max", max),
			zap.Error(err),
		)
		return nil, err
	}

	return result, nil
}

func (r *RedisClient) ClearHotspots(ctx context.Context, tripType string) error {
	var key string
	if tripType == "Inter" {
		key = InterHotspotsKey
	} else {
		key = IntraHotspotsKey
	}

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.logger.Error("failed to clear hotspots",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	r.logger.Info("cleared hotspots", zap.String("key", key))
	return nil
}

func (r *RedisClient) GetHotspotsStats(ctx context.Context, tripType string) (map[string]interface{}, error) {
	var key string
	if tripType == "Inter" {
		key = InterHotspotsKey
	} else {
		key = IntraHotspotsKey
	}

	pipe := r.client.Pipeline()
	countCmd := pipe.ZCard(ctx, key)
	ttlCmd := pipe.TTL(ctx, key)
	topCmd := pipe.ZRevRangeWithScores(ctx, key, 0, 0)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		r.logger.Error("failed to get hotspots stats",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}

	stats := map[string]interface{}{
		"total_locations": countCmd.Val(),
		"ttl_seconds":     ttlCmd.Val().Seconds(),
		"trip_type":       tripType,
	}

	if topLocations := topCmd.Val(); len(topLocations) > 0 {
		stats["top_location"] = map[string]interface{}{
			"location": topLocations[0].Member,
			"score":    topLocations[0].Score,
		}
	}

	return stats, nil
}

func (r *RedisClient) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
