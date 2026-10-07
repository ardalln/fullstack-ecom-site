package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"shop-api/internal/domain"
)

type TicketRepository struct{ db DBTX }

func NewTicketRepository(db DBTX) *TicketRepository { return &TicketRepository{db: db} }

const ticketColumns = `t.id, t.user_id, concat(u.first_name, ' ', u.last_name), u.phone, t.subject, t.status,
	COALESCE(last_message.body, ''), t.created_at, t.updated_at`
const ticketFrom = ` FROM support_tickets t JOIN users u ON u.id=t.user_id
	LEFT JOIN LATERAL (SELECT body FROM support_ticket_messages WHERE ticket_id=t.id ORDER BY id DESC LIMIT 1) last_message ON TRUE `

func scanTicket(row pgx.Row) (*domain.SupportTicket, error) {
	var ticket domain.SupportTicket
	var status string
	if err := row.Scan(&ticket.ID, &ticket.UserID, &ticket.UserName, &ticket.UserPhone, &ticket.Subject, &status,
		&ticket.LastMessage, &ticket.CreatedAt, &ticket.UpdatedAt); err != nil {
		return nil, err
	}
	ticket.Status = domain.TicketStatus(status)
	return &ticket, nil
}

func (r *TicketRepository) Create(ctx context.Context, ticket *domain.SupportTicket) error {
	err := r.db.QueryRow(ctx, `INSERT INTO support_tickets (user_id, subject, status) VALUES ($1,$2,$3)
		RETURNING id, created_at, updated_at`, ticket.UserID, ticket.Subject, ticket.Status).
		Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create support ticket: %w", err)
	}
	return nil
}

func (r *TicketRepository) GetByID(ctx context.Context, id int64) (*domain.SupportTicket, error) {
	return r.get(ctx, id, false)
}

func (r *TicketRepository) GetByIDForUpdate(ctx context.Context, id int64) (*domain.SupportTicket, error) {
	return r.get(ctx, id, true)
}

func (r *TicketRepository) get(ctx context.Context, id int64, lock bool) (*domain.SupportTicket, error) {
	query := `SELECT ` + ticketColumns + ticketFrom + `WHERE t.id=$1`
	if lock {
		query += ` FOR UPDATE OF t`
	}
	ticket, err := scanTicket(r.db.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get support ticket: %w", err)
	}
	return ticket, nil
}

func (r *TicketRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.SupportTicket, int, error) {
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM support_tickets WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count user support tickets: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+ticketColumns+ticketFrom+`WHERE t.user_id=$1 ORDER BY t.updated_at DESC, t.id DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list user support tickets: %w", err)
	}
	defer rows.Close()
	tickets := make([]domain.SupportTicket, 0, limit)
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan support ticket: %w", err)
		}
		tickets = append(tickets, *ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate user support tickets: %w", err)
	}
	return tickets, total, nil
}

func (r *TicketRepository) ListAll(ctx context.Context, status domain.TicketStatus, limit, offset int) ([]domain.SupportTicket, int, error) {
	var filter any
	if status != "" {
		filter = string(status)
	}
	const where = `$1::text IS NULL OR t.status=$1`
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM support_tickets t WHERE `+where, filter).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count support tickets: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+ticketColumns+ticketFrom+`WHERE `+where+` ORDER BY t.updated_at DESC, t.id DESC LIMIT $2 OFFSET $3`, filter, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list support tickets: %w", err)
	}
	defer rows.Close()
	tickets := make([]domain.SupportTicket, 0, limit)
	for rows.Next() {
		ticket, err := scanTicket(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("scan support ticket: %w", err)
		}
		tickets = append(tickets, *ticket)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate support tickets: %w", err)
	}
	return tickets, total, nil
}

func (r *TicketRepository) ListMessages(ctx context.Context, ticketID int64) ([]domain.TicketMessage, error) {
	rows, err := r.db.Query(ctx, `SELECT id, ticket_id, sender_id, sender_name, sender_role, body, created_at
		FROM support_ticket_messages WHERE ticket_id=$1 ORDER BY id`, ticketID)
	if err != nil {
		return nil, fmt.Errorf("list support ticket messages: %w", err)
	}
	defer rows.Close()
	messages := make([]domain.TicketMessage, 0)
	for rows.Next() {
		var message domain.TicketMessage
		var role string
		if err := rows.Scan(&message.ID, &message.TicketID, &message.SenderID, &message.SenderName, &role, &message.Body, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan support ticket message: %w", err)
		}
		message.SenderRole = domain.Role(role)
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate support ticket messages: %w", err)
	}
	return messages, nil
}

func (r *TicketRepository) AddMessage(ctx context.Context, message *domain.TicketMessage) error {
	err := r.db.QueryRow(ctx, `INSERT INTO support_ticket_messages (ticket_id, sender_id, sender_name, sender_role, body)
		VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`, message.TicketID, message.SenderID, message.SenderName,
		message.SenderRole, message.Body).Scan(&message.ID, &message.CreatedAt)
	if err != nil {
		return fmt.Errorf("add support ticket message: %w", err)
	}
	return nil
}

func (r *TicketRepository) SetStatus(ctx context.Context, id int64, status domain.TicketStatus) error {
	tag, err := r.db.Exec(ctx, `UPDATE support_tickets SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	if err != nil {
		return fmt.Errorf("set support ticket status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
