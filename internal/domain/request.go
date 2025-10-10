package domain

import (
	"context"
	"time"
)

type Request struct {
	UserName     Username     `json:"user_name"`
	MobileNumber MobileNumber `json:"mobile_number"`
	TripType     string       `json:"trip_type"`
	OriginLat    float64      `json:"origin_lat"`
	OriginLong   float64      `json:"origin_long"`
	OriginIndex  int64        `json:"origin_index"`
	DestLat      float64      `json:"dest_lat"`
	DestLong     float64      `json:"dest_long"`
	DestIndex    int64        `json:"dest_index"`
	Distance     float64      `json:"distance"`
	CreatedAt    time.Time    `json:"created_at"`
}

type RequestService interface {
	CreateByInterType(context.Context, Request) error
	CreateByIntraType(context.Context, Request) error
	GetByMobileNumber(context.Context, MobileNumber) ([]Request, error)
	GetByUserName(context.Context, Username) ([]Request, error)
	GetByIndex(context.Context, int64) ([]Request, error)
	Delete(context.Context, MobileNumber) error
	GetHotspots(context.Context) (map[string][]string, error)
}

type RequestStorage interface {
	Store(context.Context, Request) error
	FindByMobileNumber(context.Context, MobileNumber) ([]Request, error)
	FindByUserName(context.Context, Username) ([]Request, error)
	FindByIndex(context.Context, int64) ([]Request, error)
	Delete(context.Context, MobileNumber) error
}
