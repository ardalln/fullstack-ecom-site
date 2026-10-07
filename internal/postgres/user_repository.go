package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"shop-api/internal/domain"
)

type UserRepository struct {
	db DBTX
}

func NewUserRepository(db DBTX) *UserRepository {
	return &UserRepository{db: db}
}

const userColumns = `id, first_name, last_name, username, phone, email, role, is_active, created_at`

func scanUser(row pgx.Row) (*domain.User, error) {
	var (
		u    domain.User
		role string
	)
	if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Username, &u.Phone, &u.Email, &role, &u.IsActive, &u.CreatedAt); err != nil {
		return nil, err
	}
	u.Role = domain.Role(role)
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	const q = `
		INSERT INTO users (first_name, last_name, username, phone, email, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, q, u.FirstName, u.LastName, u.Username, u.Phone, u.Email, string(u.Role)).
		Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		if hasPgCode(err, pgUniqueViolation) {
			if name, ok := pgConstraint(err); ok {
				switch name {
				case "users_email_key":
					return domain.ErrEmailTaken
				case "users_username_lower_key", "users_username_key":
					return domain.ErrUsernameTaken
				}
			}
			return domain.ErrPhoneTaken
		}
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE phone = $1`, phone))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	return u, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	u, err := scanUser(r.db.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE username = lower($1)`, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

func (r *UserRepository) List(ctx context.Context, search string, limit, offset int) ([]domain.User, int, error) {
	pattern := "%" + search + "%"
	const countQuery = `SELECT COUNT(*) FROM users
		WHERE $1 = '' OR first_name ILIKE $2 OR last_name ILIKE $2 OR username ILIKE $2 OR phone ILIKE $2 OR COALESCE(email, '') ILIKE $2`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, search, pattern).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+userColumns+` FROM users
		WHERE $1 = '' OR first_name ILIKE $2 OR last_name ILIKE $2 OR username ILIKE $2 OR phone ILIKE $2 OR COALESCE(email, '') ILIKE $2
		ORDER BY id DESC LIMIT $3 OFFSET $4`, search, pattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0, limit)
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users: %w", err)
	}
	return users, total, nil
}

func (r *UserRepository) UpdateAccess(ctx context.Context, id int64, role domain.Role, active bool) error {
	// This method is called inside a database transaction. Locking every current
	// active admin serializes concurrent changes and protects the last admin.
	rows, err := r.db.Query(ctx, `SELECT id FROM users WHERE role = 'admin' AND is_active ORDER BY id FOR UPDATE`)
	if err != nil {
		return fmt.Errorf("lock active admins: %w", err)
	}
	activeAdminIDs := make(map[int64]struct{})
	for rows.Next() {
		var adminID int64
		if err := rows.Scan(&adminID); err != nil {
			rows.Close()
			return fmt.Errorf("scan active admin: %w", err)
		}
		activeAdminIDs[adminID] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate active admins: %w", err)
	}
	rows.Close()

	if _, isActiveAdmin := activeAdminIDs[id]; isActiveAdmin && (role != domain.RoleAdmin || !active) && len(activeAdminIDs) == 1 {
		return domain.ErrLastAdmin
	}
	tag, err := r.db.Exec(ctx, `UPDATE users SET role = $2, is_active = $3 WHERE id = $1`, id, string(role), active)
	if err != nil {
		return fmt.Errorf("update user access: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
