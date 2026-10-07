package handler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/labstack/echo/v4"

	"shop-api/internal/domain"
)

var analyticsUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-(?:[0-9a-fA-F]{4}-){3}[0-9a-fA-F]{12}$`)

type AnalyticsHandler struct{ repository domain.AnalyticsRepository }

func NewAnalyticsHandler(repository domain.AnalyticsRepository) *AnalyticsHandler {
	return &AnalyticsHandler{repository: repository}
}

type AnalyticsHeartbeatRequest struct {
	VisitorID string `json:"visitor_id" validate:"required,len=36"`
	Path      string `json:"path" validate:"required,max=180"`
}

type AnalyticsViewRequest struct {
	VisitorID string `json:"visitor_id" validate:"required,len=36"`
	ViewID    string `json:"view_id" validate:"required,len=36"`
	Path      string `json:"path" validate:"required,max=180"`
}

type AnalyticsVisitorResponse struct {
	VisitorID string    `json:"visitor_id"`
	Path      string    `json:"path"`
	LastSeen  time.Time `json:"last_seen"`
}

type AnalyticsOverviewResponse struct {
	ActiveOnline   int64                      `json:"active_online"`
	VisitorsToday  int64                      `json:"visitors_today"`
	UniqueVisitors int64                      `json:"unique_visitors"`
	PageViewsToday int64                      `json:"page_views_today"`
	ActiveVisitors []AnalyticsVisitorResponse `json:"active_visitors"`
}

func (h *AnalyticsHandler) Heartbeat(c echo.Context) error {
	var request AnalyticsHeartbeatRequest
	if err := bindAndValidate(c, &request); err != nil {
		return err
	}
	if !analyticsUUID.MatchString(request.VisitorID) {
		return fmt.Errorf("%w: visitor_id must be a UUID", domain.ErrInvalidInput)
	}
	if err := validateAnalyticsPath(request.Path); err != nil {
		return err
	}
	if err := h.repository.Heartbeat(c.Request().Context(), request.VisitorID, request.Path); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AnalyticsHandler) RecordView(c echo.Context) error {
	var request AnalyticsViewRequest
	if err := bindAndValidate(c, &request); err != nil {
		return err
	}
	if !analyticsUUID.MatchString(request.VisitorID) || !analyticsUUID.MatchString(request.ViewID) {
		return fmt.Errorf("%w: analytics identifiers must be UUIDs", domain.ErrInvalidInput)
	}
	if err := validateAnalyticsPath(request.Path); err != nil {
		return err
	}
	if err := h.repository.RecordPageView(c.Request().Context(), request.ViewID, request.VisitorID, request.Path); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *AnalyticsHandler) Overview(c echo.Context) error {
	data, err := h.repository.GetAnalyticsOverview(c.Request().Context())
	if err != nil {
		return err
	}
	response := AnalyticsOverviewResponse{
		ActiveOnline: data.ActiveOnline, VisitorsToday: data.VisitorsToday,
		UniqueVisitors: data.UniqueVisitors, PageViewsToday: data.PageViewsToday,
		ActiveVisitors: make([]AnalyticsVisitorResponse, 0, len(data.ActiveVisitors)),
	}
	for _, visitor := range data.ActiveVisitors {
		response.ActiveVisitors = append(response.ActiveVisitors, AnalyticsVisitorResponse{
			VisitorID: visitor.VisitorID, Path: visitor.Path, LastSeen: visitor.LastSeen,
		})
	}
	return c.JSON(http.StatusOK, response)
}

func validateAnalyticsPath(path string) error {
	if path == "/" {
		return nil
	}
	if len([]rune(path)) > 180 || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "?#\\") {
		return fmt.Errorf("%w: analytics path is invalid", domain.ErrInvalidInput)
	}
	for _, char := range path {
		if unicode.IsControl(char) || unicode.IsSpace(char) {
			return fmt.Errorf("%w: analytics path is invalid", domain.ErrInvalidInput)
		}
	}
	for _, segment := range strings.Split(path, "/") {
		if segment == ".." || segment == "." {
			return fmt.Errorf("%w: analytics path is invalid", domain.ErrInvalidInput)
		}
	}
	return nil
}
