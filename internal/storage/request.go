package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/hotspot/internal/domain"
)

type request struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewRequestStore(pool *pgxpool.Pool, logger *zap.Logger) domain.RequestStorage {
	return &request{pool: pool, logger: logger}
}

func (s *request) Store(ctx context.Context, req domain.Request) error {
	const query = `
		INSERT INTO requests (
			user_name, mobile_number, trip_type, origin_lat, origin_long, 
			origin_index, dest_lat, dest_long, dest_index, distance, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := s.pool.Exec(ctx, query,
		string(req.UserName),
		string(req.MobileNumber),
		req.TripType,
		req.OriginLat,
		req.OriginLong,
		req.OriginIndex,
		req.DestLat,
		req.DestLong,
		req.DestIndex,
		req.Distance,
		req.CreatedAt,
	)

	if err != nil {
		s.logger.Error("failed to store request",
			zap.String("user_name", string(req.UserName)),
			zap.String("mobile_number", string(req.MobileNumber)),
			zap.String("trip_type", req.TripType),
			zap.Error(err),
		)
		return fmt.Errorf("failed to store request: %w", err)
	}

	s.logger.Debug("stored request successfully",
		zap.String("user_name", string(req.UserName)),
		zap.String("trip_type", req.TripType),
		zap.Float64("distance", req.Distance),
	)

	return nil
}

func (s *request) FindByMobileNumber(ctx context.Context, mobile domain.MobileNumber) ([]domain.Request, error) {
	const query = `
		SELECT user_name, mobile_number, trip_type, origin_lat, origin_long, 
		       origin_index, dest_lat, dest_long, dest_index, distance, created_at
		FROM requests 
		WHERE mobile_number = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, string(mobile))
	if err != nil {
		s.logger.Error("failed to query requests by mobile number",
			zap.String("mobile_number", string(mobile)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to query requests by mobile number: %w", err)
	}
	defer rows.Close()

	var requests []domain.Request
	for rows.Next() {
		var req domain.Request
		var userName, mobileNumber string
		
		err := rows.Scan(
			&userName,
			&mobileNumber,
			&req.TripType,
			&req.OriginLat,
			&req.OriginLong,
			&req.OriginIndex,
			&req.DestLat,
			&req.DestLong,
			&req.DestIndex,
			&req.Distance,
			&req.CreatedAt,
		)
		if err != nil {
			s.logger.Error("failed to scan request row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan request row: %w", err)
		}

		req.UserName = domain.Username(userName)
		req.MobileNumber = domain.MobileNumber(mobileNumber)
		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("error iterating request rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating request rows: %w", err)
	}

	return requests, nil
}

func (s *request) FindByUserName(ctx context.Context, username domain.Username) ([]domain.Request, error) {
	const query = `
		SELECT user_name, mobile_number, trip_type, origin_lat, origin_long, 
		       origin_index, dest_lat, dest_long, dest_index, distance, created_at
		FROM requests 
		WHERE user_name = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, string(username))
	if err != nil {
		s.logger.Error("failed to query requests by username",
			zap.String("username", string(username)),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to query requests by username: %w", err)
	}
	defer rows.Close()

	var requests []domain.Request
	for rows.Next() {
		var req domain.Request
		var userName, mobileNumber string
		
		err := rows.Scan(
			&userName,
			&mobileNumber,
			&req.TripType,
			&req.OriginLat,
			&req.OriginLong,
			&req.OriginIndex,
			&req.DestLat,
			&req.DestLong,
			&req.DestIndex,
			&req.Distance,
			&req.CreatedAt,
		)
		if err != nil {
			s.logger.Error("failed to scan request row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan request row: %w", err)
		}

		req.UserName = domain.Username(userName)
		req.MobileNumber = domain.MobileNumber(mobileNumber)
		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("error iterating request rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating request rows: %w", err)
	}

	return requests, nil
}

func (s *request) FindByIndex(ctx context.Context, index int64) ([]domain.Request, error) {
	const query = `
		SELECT user_name, mobile_number, trip_type, origin_lat, origin_long, 
		       origin_index, dest_lat, dest_long, dest_index, distance, created_at
		FROM requests 
		WHERE origin_index = $1 OR dest_index = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, index)
	if err != nil {
		s.logger.Error("failed to query requests by H3 index",
			zap.Int64("index", index),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to query requests by H3 index: %w", err)
	}
	defer rows.Close()

	var requests []domain.Request
	for rows.Next() {
		var req domain.Request
		var userName, mobileNumber string
		
		err := rows.Scan(
			&userName,
			&mobileNumber,
			&req.TripType,
			&req.OriginLat,
			&req.OriginLong,
			&req.OriginIndex,
			&req.DestLat,
			&req.DestLong,
			&req.DestIndex,
			&req.Distance,
			&req.CreatedAt,
		)
		if err != nil {
			s.logger.Error("failed to scan request row", zap.Error(err))
			return nil, fmt.Errorf("failed to scan request row: %w", err)
		}

		req.UserName = domain.Username(userName)
		req.MobileNumber = domain.MobileNumber(mobileNumber)
		requests = append(requests, req)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("error iterating request rows", zap.Error(err))
		return nil, fmt.Errorf("error iterating request rows: %w", err)
	}

	return requests, nil
}

func (s *request) Delete(ctx context.Context, mobile domain.MobileNumber) error {
	const query = `DELETE FROM requests WHERE mobile_number = $1`

	result, err := s.pool.Exec(ctx, query, string(mobile))
	if err != nil {
		s.logger.Error("failed to delete requests by mobile number",
			zap.String("mobile_number", string(mobile)),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete requests by mobile number: %w", err)
	}

	rowsAffected := result.RowsAffected()
	s.logger.Info("deleted requests",
		zap.String("mobile_number", string(mobile)),
		zap.Int64("rows_affected", rowsAffected),
	)

	return nil
}
