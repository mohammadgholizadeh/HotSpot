package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"github.com/mohammadgholizadeh/hotspot/internal/broker"
	"github.com/mohammadgholizadeh/hotspot/internal/domain"
)

type requestHandler struct {
	svc      domain.RequestService
	logger   *zap.Logger
	rabbitMQ *broker.RabbitMQClient
}

func RequestRoutes(g *echo.Group, svc domain.RequestService, logger *zap.Logger, rabbitMQ *broker.RabbitMQClient) {
	h := &requestHandler{
		svc:      svc,
		logger:   logger,
		rabbitMQ: rabbitMQ,
	}
	g.POST("/requests", h.create)
	g.GET("/requests/mobile", h.getByMobileNumber)
	g.GET("/requests/username", h.getByUsername)
	g.GET("/requests/index", h.getByIndex)
	g.DELETE("/requests", h.delete)
	g.GET("/hotspots", h.getHotspots)
}

func (h *requestHandler) create(c echo.Context) error {
	type createReq struct {
		UserName     domain.Username     `json:"user_name"`
		MobileNumber domain.MobileNumber `json:"mobile_number"`
		TripType     string              `json:"trip_type"`
		OriginLat    float64             `json:"origin_lat"`
		OriginLong   float64             `json:"origin_long"`
		DestLat      float64             `json:"dest_lat"`
		DestLong     float64             `json:"dest_long"`
	}

	var req createReq
	if err := c.Bind(&req); err != nil {
		h.logger.Warn("request.create: invalid request body", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.TripType != "Intra" && req.TripType != "Inter" {
		h.logger.Warn("request.create: invalid trip type", zap.String("trip_type", req.TripType))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "trip_type must be 'Intra' or 'Inter'"})
	}

	request := domain.Request{
		UserName:     req.UserName,
		MobileNumber: req.MobileNumber,
		TripType:     req.TripType,
		OriginLat:    req.OriginLat,
		OriginLong:   req.OriginLong,
		DestLat:      req.DestLat,
		DestLong:     req.DestLong,
		CreatedAt:    time.Now(),
	}

	if err := h.publishToQueue(request); err != nil {
		h.logger.Error("request.create: failed to publish to queue", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue request"})
	}

	hotspots, err := h.svc.GetHotspots(c.Request().Context())
	if err != nil {
		h.logger.Warn("request.create: failed to fetch hotspots snapshot", zap.Error(err))
		return c.JSON(http.StatusAccepted, map[string]interface{}{
			"accepted": true,
			"message":  "request enqueued; hotspots unavailable",
		})
	}

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"accepted": true,
		"hotspots": hotspots,
	})
}

func (h *requestHandler) publishToQueue(request domain.Request) error {
	ch, err := h.rabbitMQ.NewChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	body, err := json.Marshal(request)
	if err != nil {
		h.logger.Error("failed to marshal request", zap.Error(err))
		return err
	}

	var routingKey string
	if request.TripType == "Intra" {
		routingKey = broker.KeyIntra
	} else {
		routingKey = broker.KeyInter
	}

	err = ch.Publish(
		broker.ExchangeRideType,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})

	if err != nil {
		h.logger.Error("failed to publish message",
			zap.String("routing_key", routingKey),
			zap.Error(err))
		return err
	}

	h.logger.Info("published message to queue",
		zap.String("trip_type", request.TripType),
		zap.String("routing_key", routingKey))

	return nil
}

func (h *requestHandler) getByMobileNumber(c echo.Context) error {
	mobile := domain.MobileNumber(c.QueryParam("mobile"))
	if err := mobile.Validate(); err != nil {
		h.logger.Warn("request.get_by_mobile: invalid mobile", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid mobile number"})
	}

	requests, err := h.svc.GetByMobileNumber(c.Request().Context(), mobile)
	if err != nil {
		h.logger.Error("request.get_by_mobile failed", zap.String("mobile", string(mobile)), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, requests)
}

func (h *requestHandler) getByUsername(c echo.Context) error {
	username := domain.Username(c.QueryParam("username"))
	if err := username.Validate(); err != nil {
		h.logger.Warn("request.get_by_username: invalid username", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid username"})
	}

	requests, err := h.svc.GetByUserName(c.Request().Context(), username)
	if err != nil {
		h.logger.Error("request.get_by_username failed", zap.String("username", string(username)), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, requests)
}

func (h *requestHandler) getByIndex(c echo.Context) error {
	indexParam := c.QueryParam("index")
	if indexParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "index parameter is required"})
	}

	var index int64
	if _, err := fmt.Sscanf(indexParam, "%d", &index); err != nil {
		h.logger.Warn("request.get_by_index: invalid index", zap.String("index", indexParam), zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid index format"})
	}

	requests, err := h.svc.GetByIndex(c.Request().Context(), index)
	if err != nil {
		h.logger.Error("request.get_by_index failed", zap.Int64("index", index), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, requests)
}

func (h *requestHandler) delete(c echo.Context) error {
	mobile := domain.MobileNumber(c.QueryParam("mobile"))
	if err := mobile.Validate(); err != nil {
		h.logger.Warn("request.delete: invalid mobile", zap.Error(err))
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid mobile number"})
	}

	if err := h.svc.Delete(c.Request().Context(), mobile); err != nil {
		h.logger.Error("request.delete failed", zap.String("mobile", string(mobile)), zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]bool{"success": true})
}

func (h *requestHandler) getHotspots(c echo.Context) error {
	hotspots, err := h.svc.GetHotspots(c.Request().Context())
	if err != nil {
		h.logger.Error("request.get_hotspots failed", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, hotspots)
}
