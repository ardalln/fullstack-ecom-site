package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"shop-api/internal/domain"
	"shop-api/internal/service"
)

type CommentHandler struct{ svc *service.CommentService }

func NewCommentHandler(svc *service.CommentService) *CommentHandler { return &CommentHandler{svc: svc} }

func (h *CommentHandler) ListApproved(c echo.Context) error {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		return domain.ErrNotFound
	}
	comments, err := h.svc.ListApproved(c.Request().Context(), productID)
	if err != nil {
		return err
	}
	data := make([]ProductCommentResponse, 0, len(comments))
	for i := range comments {
		data = append(data, toProductCommentResponse(&comments[i]))
	}
	return c.JSON(http.StatusOK, data)
}

func (h *CommentHandler) Create(c echo.Context) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		return domain.ErrNotFound
	}
	var req ProductCommentRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	comment, err := h.svc.Create(c.Request().Context(), productID, userID, req.Body)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toProductCommentResponse(comment))
}

func (h *CommentHandler) ListAll(c echo.Context) error {
	page, limit := pagination(c)
	var approved *bool
	switch c.QueryParam("status") {
	case "":
	case "pending":
		value := false
		approved = &value
	case "approved":
		value := true
		approved = &value
	default:
		return domain.ErrInvalidInput
	}
	comments, total, err := h.svc.ListAll(c.Request().Context(), approved, page, limit)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, toProductCommentPage(comments, page, limit, total))
}

func (h *CommentHandler) SetApproval(c echo.Context) error {
	id, err := idParam(c)
	if err != nil {
		return err
	}
	var req CommentApprovalRequest
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}
	if err := h.svc.SetApproval(c.Request().Context(), id, *req.IsApproved); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}
