package service

import (
	"context"
	"fmt"
	"strings"

	"shop-api/internal/domain"
)

type CommentService struct {
	comments domain.CommentRepository
	products domain.ProductRepository
}

func NewCommentService(comments domain.CommentRepository, products domain.ProductRepository) *CommentService {
	return &CommentService{comments: comments, products: products}
}

func (s *CommentService) Create(ctx context.Context, productID, userID int64, body string) (*domain.ProductComment, error) {
	body = strings.TrimSpace(body)
	if productID <= 0 || userID <= 0 || body == "" || len([]rune(body)) > 2000 {
		return nil, fmt.Errorf("%w: comment must contain 1-2000 characters", domain.ErrInvalidInput)
	}
	if _, err := s.products.GetByID(ctx, productID); err != nil {
		return nil, err
	}
	comment := &domain.ProductComment{ProductID: productID, UserID: userID, Body: body}
	if err := s.comments.Create(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *CommentService) ListApproved(ctx context.Context, productID int64) ([]domain.ProductComment, error) {
	if productID <= 0 {
		return nil, domain.ErrNotFound
	}
	if _, err := s.products.GetByID(ctx, productID); err != nil {
		return nil, err
	}
	return s.comments.ListApprovedByProduct(ctx, productID)
}

func (s *CommentService) ListAll(ctx context.Context, approved *bool, page, limit int) ([]domain.ProductComment, int, error) {
	return s.comments.ListAll(ctx, approved, limit, (page-1)*limit)
}

func (s *CommentService) SetApproval(ctx context.Context, id int64, approved bool) error {
	if id <= 0 {
		return domain.ErrNotFound
	}
	return s.comments.SetApproval(ctx, id, approved)
}
