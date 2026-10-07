package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"shop-api/internal/domain"
)

type CommentRepository struct{ db DBTX }

func NewCommentRepository(db DBTX) *CommentRepository { return &CommentRepository{db: db} }

const commentColumns = `c.id, c.product_id, p.name, p.slug, c.user_id,
	concat(u.first_name, ' ', u.last_name), c.body, c.is_approved, c.created_at, c.updated_at`
const commentFrom = ` FROM product_comments c JOIN products p ON p.id = c.product_id JOIN users u ON u.id = c.user_id `

func scanComment(row pgx.Row) (*domain.ProductComment, error) {
	var comment domain.ProductComment
	err := row.Scan(&comment.ID, &comment.ProductID, &comment.ProductName, &comment.ProductSlug, &comment.UserID,
		&comment.AuthorName, &comment.Body, &comment.IsApproved, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) Create(ctx context.Context, comment *domain.ProductComment) error {
	err := r.db.QueryRow(ctx, `INSERT INTO product_comments (product_id, user_id, body)
		VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`, comment.ProductID, comment.UserID, comment.Body).
		Scan(&comment.ID, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create product comment: %w", err)
	}
	return nil
}

func (r *CommentRepository) ListApprovedByProduct(ctx context.Context, productID int64) ([]domain.ProductComment, error) {
	rows, err := r.db.Query(ctx, `SELECT `+commentColumns+commentFrom+
		`WHERE c.product_id = $1 AND c.is_approved ORDER BY c.created_at DESC, c.id DESC LIMIT 100`, productID)
	if err != nil {
		return nil, fmt.Errorf("list approved product comments: %w", err)
	}
	defer rows.Close()
	comments := make([]domain.ProductComment, 0)
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return nil, fmt.Errorf("scan product comment: %w", err)
		}
		comments = append(comments, *comment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate product comments: %w", err)
	}
	return comments, nil
}

func (r *CommentRepository) ListAll(ctx context.Context, approved *bool, limit, offset int) ([]domain.ProductComment, int, error) {
	var filter any
	if approved != nil {
		filter = *approved
	}
	const where = `($1::boolean IS NULL OR c.is_approved = $1)`
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM product_comments c WHERE `+where, filter).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count product comments: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+commentColumns+commentFrom+`WHERE `+where+
		` ORDER BY c.created_at DESC, c.id DESC LIMIT $2 OFFSET $3`, filter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list product comments: %w", err)
	}
	defer rows.Close()
	comments := make([]domain.ProductComment, 0, limit)
	for rows.Next() {
		comment, err := scanComment(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan product comment: %w", err)
		}
		comments = append(comments, *comment)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate product comments: %w", err)
	}
	return comments, total, nil
}

func (r *CommentRepository) SetApproval(ctx context.Context, id int64, approved bool) error {
	tag, err := r.db.Exec(ctx, `UPDATE product_comments SET is_approved = $2, updated_at = now() WHERE id = $1`, id, approved)
	if err != nil {
		return fmt.Errorf("set product comment approval: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
