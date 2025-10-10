package service

import (
	"context"
	"fmt"
	"math"

	"github.com/redis/go-redis/v9"
	"github.com/uber/h3-go/v4"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/hotspot/internal/cache"
	"github.com/mohammadgholizadeh/hotspot/internal/domain"
)

type request struct {
	store  domain.RequestStorage
	redis  *redis.Client
	logger *zap.Logger
}

const (
	H3Resolution  = 7
	EarthRadiusKm = 6371.0
)

func NewRequestService(store domain.RequestStorage, redis *redis.Client, logger *zap.Logger) domain.RequestService {
	return &request{
		store:  store,
		redis:  redis,
		logger: logger,
	}
}

func (s *request) CreateByInterType(ctx context.Context, req domain.Request) error {
	return s.processRequest(ctx, req, true)
}
func (s *request) CreateByIntraType(ctx context.Context, req domain.Request) error {
	return s.processRequest(ctx, req, false)
}

func (s *request) processRequest(ctx context.Context, req domain.Request, isInter bool) error {
	// Calculate H3 indices for origin and destination
	originLatLng := h3.NewLatLng(req.OriginLat, req.OriginLong)
	destLatLng := h3.NewLatLng(req.DestLat, req.DestLong)

	originIndex := h3.LatLngToCell(originLatLng, H3Resolution)
	destIndex := h3.LatLngToCell(destLatLng, H3Resolution)

	req.OriginIndex = int64(originIndex)
	req.DestIndex = int64(destIndex)

	distance := s.calculateDistance(req.OriginLat, req.OriginLong, req.DestLat, req.DestLong)
	req.Distance = distance

	s.logger.Info("processed trip request",
		zap.String("trip_type", req.TripType),
		zap.Float64("distance", distance),
		zap.Int64("origin_index", req.OriginIndex),
		zap.Int64("dest_index", req.DestIndex),
	)

	if err := s.store.Store(ctx, req); err != nil {
		s.logger.Error("failed to store request", zap.Error(err))
		return fmt.Errorf("failed to store request: %w", err)
	}

	if err := s.updateHotspots(ctx, req, isInter); err != nil {
		s.logger.Warn("failed to update hotspots cache", zap.Error(err))
	}

	return nil
}

func (s *request) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadiusKm * c
}

func (s *request) updateHotspots(ctx context.Context, req domain.Request, isInter bool) error {
	var key string
	if isInter {
		key = cache.InterHotspotsKey
	} else {
		key = cache.IntraHotspotsKey
	}

	originKey := fmt.Sprintf("origin:%f,%f", req.OriginLat, req.OriginLong)
	destKey := fmt.Sprintf("dest:%f,%f", req.DestLat, req.DestLong)

	pipe := s.redis.Pipeline()

	pipe.ZIncrBy(ctx, key, 1, originKey)
	pipe.ZIncrBy(ctx, key, 1, destKey)

	pipe.Expire(ctx, key, cache.HotspotsTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		s.logger.Error("failed to update Redis hotspots",
			zap.String("key", key),
			zap.Error(err))
		return err
	}

	s.logger.Debug("updated hotspots cache",
		zap.String("key", key),
		zap.String("origin", originKey),
		zap.String("dest", destKey))

	return nil
}

func (s *request) GetByMobileNumber(ctx context.Context, mobile domain.MobileNumber) ([]domain.Request, error) {
	if err := mobile.Validate(); err != nil {
		s.logger.Warn("invalid mobile number", zap.Error(err))
		return nil, err
	}

	return s.store.FindByMobileNumber(ctx, mobile)
}

func (s *request) GetByUserName(ctx context.Context, username domain.Username) ([]domain.Request, error) {
	if err := username.Validate(); err != nil {
		s.logger.Warn("invalid username", zap.Error(err))
		return nil, err
	}

	return s.store.FindByUserName(ctx, username)
}

func (s *request) GetByIndex(ctx context.Context, index int64) ([]domain.Request, error) {
	if index <= 0 {
		return nil, fmt.Errorf("invalid H3 index")
	}

	return s.store.FindByIndex(ctx, index)
}

func (s *request) Delete(ctx context.Context, mobile domain.MobileNumber) error {
	if err := mobile.Validate(); err != nil {
		s.logger.Warn("invalid mobile number", zap.Error(err))
		return err
	}

	return s.store.Delete(ctx, mobile)
}

func (s *request) GetHotspots(ctx context.Context) (map[string][]string, error) {
	IntraHotspots, err := s.redis.ZRevRange(ctx, cache.IntraHotspotsKey, 0, 9).Result()
	if err != nil && err != redis.Nil {
		s.logger.Error("failed to get Intra hotspots", zap.Error(err))
		return nil, err
	}

	InterHotspots, err := s.redis.ZRevRange(ctx, cache.InterHotspotsKey, 0, 9).Result()
	if err != nil && err != redis.Nil {
		s.logger.Error("failed to get Inter hotspots", zap.Error(err))
		return nil, err
	}

	return map[string][]string{
		"Intra": IntraHotspots,
		"Inter": InterHotspots,
	}, nil
}
