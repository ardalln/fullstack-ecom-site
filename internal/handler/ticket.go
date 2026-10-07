package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"shop-api/internal/domain"
	"shop-api/internal/service"
)

type TicketHandler struct{ svc *service.TicketService }

func NewTicketHandler(svc *service.TicketService) *TicketHandler { return &TicketHandler{svc: svc} }

func (h *TicketHandler) ListMine(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	page, limit := pagination(c)
	tickets, total, err := h.svc.ListByUser(c.Request().Context(), userID, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toTicketPage(tickets, page, limit, total))
}

func (h *TicketHandler) Create(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	var req TicketCreateRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ticket, err := h.svc.Create(c.Request().Context(), userID, req.Subject, req.Body)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toTicketResponse(ticket))
}

func (h *TicketHandler) GetMine(c echo.Context) error {
	return h.get(c, false)
}

func (h *TicketHandler) GetAdmin(c echo.Context) error {
	return h.get(c, true)
}

func (h *TicketHandler) get(c echo.Context, admin bool) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	ticket, err := h.svc.Get(c.Request().Context(), id, userID, admin)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toTicketResponse(ticket))
}

func (h *TicketHandler) ReplyMine(c echo.Context) error {
	return h.reply(c, domain.RoleUser)
}

func (h *TicketHandler) ReplyAdmin(c echo.Context) error {
	return h.reply(c, domain.RoleAdmin)
}

func (h *TicketHandler) reply(c echo.Context, role domain.Role) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	var req TicketMessageRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	ticket, err := h.svc.Reply(c.Request().Context(), id, userID, role, req.Body)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toTicketResponse(ticket))
}

func (h *TicketHandler) ListAdmin(c echo.Context) error {
	page, limit := pagination(c)
	var status domain.TicketStatus
	switch strings.TrimSpace(c.QueryParam("status")) {
	case "":
	case string(domain.TicketStatusOpen), string(domain.TicketStatusAnswered), string(domain.TicketStatusClosed):
		status = domain.TicketStatus(c.QueryParam("status"))
	default:
		return fmt.Errorf("%w: unsupported ticket status", domain.ErrInvalidInput)
	}
	tickets, total, err := h.svc.ListAll(c.Request().Context(), status, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toTicketPage(tickets, page, limit, total))
}

func (h *TicketHandler) SetStatus(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req TicketStatusRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.SetStatus(c.Request().Context(), id, req.Status); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
