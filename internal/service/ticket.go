package service

import (
	"context"
	"fmt"
	"strings"

	"shop-api/internal/domain"
)

type TicketService struct {
	tickets domain.TicketRepository
	users   domain.UserRepository
	tx      domain.TxManager
}

func NewTicketService(tickets domain.TicketRepository, users domain.UserRepository, tx domain.TxManager) *TicketService {
	return &TicketService{tickets: tickets, users: users, tx: tx}
}

func (s *TicketService) Create(ctx context.Context, userID int64, subject, body string) (*domain.SupportTicket, error) {
	subject, body = strings.TrimSpace(subject), strings.TrimSpace(body)
	if userID <= 0 || subject == "" || len([]rune(subject)) > 200 || body == "" || len([]rune(body)) > 5000 {
		return nil, fmt.Errorf("%w: subject must contain 1-200 characters and message 1-5000 characters", domain.ErrInvalidInput)
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	ticket := &domain.SupportTicket{UserID: userID, UserName: strings.TrimSpace(user.FirstName + " " + user.LastName), UserPhone: user.Phone, Subject: subject, Status: domain.TicketStatusOpen}
	message := domain.TicketMessage{SenderID: user.ID, SenderName: ticket.UserName, SenderRole: domain.RoleUser, Body: body}
	err = s.tx.WithinTx(ctx, func(repos domain.Repositories) error {
		if err := repos.Tickets.Create(ctx, ticket); err != nil {
			return err
		}
		message.TicketID = ticket.ID
		return repos.Tickets.AddMessage(ctx, &message)
	})
	if err != nil {
		return nil, err
	}
	ticket.Messages = []domain.TicketMessage{message}
	return ticket, nil
}

func (s *TicketService) Get(ctx context.Context, ticketID, actorID int64, admin bool) (*domain.SupportTicket, error) {
	ticket, err := s.tickets.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if !admin && ticket.UserID != actorID {
		return nil, domain.ErrNotFound
	}
	ticket.Messages, err = s.tickets.ListMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	return ticket, nil
}

func (s *TicketService) ListByUser(ctx context.Context, userID int64, page, limit int) ([]domain.SupportTicket, int, error) {
	return s.tickets.ListByUserID(ctx, userID, limit, (page-1)*limit)
}

func (s *TicketService) ListAll(ctx context.Context, status domain.TicketStatus, page, limit int) ([]domain.SupportTicket, int, error) {
	return s.tickets.ListAll(ctx, status, limit, (page-1)*limit)
}

func (s *TicketService) Reply(ctx context.Context, ticketID, actorID int64, role domain.Role, body string) (*domain.SupportTicket, error) {
	body = strings.TrimSpace(body)
	if ticketID <= 0 || actorID <= 0 || body == "" || len([]rune(body)) > 5000 || (role != domain.RoleUser && role != domain.RoleAdmin) {
		return nil, fmt.Errorf("%w: message must contain 1-5000 characters", domain.ErrInvalidInput)
	}
	user, err := s.users.GetByID(ctx, actorID)
	if err != nil {
		return nil, err
	}
	admin := role == domain.RoleAdmin
	err = s.tx.WithinTx(ctx, func(repos domain.Repositories) error {
		ticket, err := repos.Tickets.GetByIDForUpdate(ctx, ticketID)
		if err != nil {
			return err
		}
		if !admin && ticket.UserID != actorID {
			return domain.ErrNotFound
		}
		if ticket.Status == domain.TicketStatusClosed {
			return domain.ErrTicketClosed
		}
		message := &domain.TicketMessage{
			TicketID: ticketID, SenderID: user.ID,
			SenderName: strings.TrimSpace(user.FirstName + " " + user.LastName),
			SenderRole: role, Body: body,
		}
		if err := repos.Tickets.AddMessage(ctx, message); err != nil {
			return err
		}
		status := domain.TicketStatusOpen
		if admin {
			status = domain.TicketStatusAnswered
		}
		return repos.Tickets.SetStatus(ctx, ticketID, status)
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, ticketID, actorID, admin)
}

func (s *TicketService) SetStatus(ctx context.Context, ticketID int64, status domain.TicketStatus) error {
	if ticketID <= 0 || (status != domain.TicketStatusOpen && status != domain.TicketStatusAnswered && status != domain.TicketStatusClosed) {
		return domain.ErrInvalidInput
	}
	return s.tickets.SetStatus(ctx, ticketID, status)
}
