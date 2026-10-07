package handler

import (
	"time"

	"shop-api/internal/domain"
	"shop-api/internal/richtext"
)

type BlogPostRequest struct {
	Title          string `json:"title" validate:"required,max=200"`
	Slug           string `json:"slug" validate:"omitempty,max=180"`
	Summary        string `json:"summary" validate:"max=500"`
	Content        string `json:"content" validate:"required,max=200000"`
	CoverImageURL  string `json:"cover_image_url" validate:"omitempty,max=1000"`
	SEOTitle       string `json:"seo_title" validate:"omitempty,max=200"`
	SEODescription string `json:"seo_description" validate:"omitempty,max=320"`
	IsPublished    bool   `json:"is_published"`
}

type BlogPostResponse struct {
	ID             int64      `json:"id"`
	Title          string     `json:"title"`
	Slug           string     `json:"slug"`
	Summary        string     `json:"summary"`
	Content        string     `json:"content,omitempty"`
	CoverImageURL  string     `json:"cover_image_url,omitempty"`
	SEOTitle       string     `json:"seo_title"`
	SEODescription string     `json:"seo_description"`
	IsPublished    bool       `json:"is_published"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TicketCreateRequest struct {
	Subject string `json:"subject" validate:"required,max=200"`
	Body    string `json:"body" validate:"required,max=5000"`
}

type TicketMessageRequest struct {
	Body string `json:"body" validate:"required,max=5000"`
}

type TicketStatusRequest struct {
	Status domain.TicketStatus `json:"status" validate:"required,oneof=open answered closed"`
}

type TicketMessageResponse struct {
	ID         int64       `json:"id"`
	SenderName string      `json:"sender_name"`
	SenderRole domain.Role `json:"sender_role"`
	Body       string      `json:"body"`
	CreatedAt  time.Time   `json:"created_at"`
}

type TicketResponse struct {
	ID          int64                   `json:"id"`
	UserID      int64                   `json:"user_id,omitempty"`
	UserName    string                  `json:"user_name,omitempty"`
	UserPhone   string                  `json:"user_phone,omitempty"`
	Subject     string                  `json:"subject"`
	Status      domain.TicketStatus     `json:"status"`
	LastMessage string                  `json:"last_message,omitempty"`
	Messages    []TicketMessageResponse `json:"messages,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

func toBlogPostResponse(post *domain.BlogPost) BlogPostResponse {
	return BlogPostResponse{
		ID: post.ID, Title: post.Title, Slug: post.Slug, Summary: post.Summary,
		Content: richtext.Sanitize(post.Content), CoverImageURL: post.CoverImageURL, SEOTitle: post.SEOTitle,
		SEODescription: post.SEODescription, IsPublished: post.IsPublished,
		PublishedAt: post.PublishedAt, CreatedAt: post.CreatedAt, UpdatedAt: post.UpdatedAt,
	}
}

func toBlogPostPage(posts []domain.BlogPost, page, limit, total int) PageResponse[BlogPostResponse] {
	data := make([]BlogPostResponse, 0, len(posts))
	for i := range posts {
		data = append(data, toBlogPostResponse(&posts[i]))
	}
	return PageResponse[BlogPostResponse]{Data: data, Page: page, Limit: limit, Total: total}
}

func toTicketResponse(ticket *domain.SupportTicket) TicketResponse {
	response := TicketResponse{
		ID: ticket.ID, UserID: ticket.UserID, UserName: ticket.UserName, UserPhone: ticket.UserPhone,
		Subject: ticket.Subject, Status: ticket.Status, LastMessage: ticket.LastMessage,
		CreatedAt: ticket.CreatedAt, UpdatedAt: ticket.UpdatedAt,
	}
	if ticket.Messages != nil {
		response.Messages = make([]TicketMessageResponse, 0, len(ticket.Messages))
		for _, message := range ticket.Messages {
			response.Messages = append(response.Messages, TicketMessageResponse{
				ID: message.ID, SenderName: message.SenderName, SenderRole: message.SenderRole,
				Body: message.Body, CreatedAt: message.CreatedAt,
			})
		}
	}
	return response
}

func toTicketPage(tickets []domain.SupportTicket, page, limit, total int) PageResponse[TicketResponse] {
	data := make([]TicketResponse, 0, len(tickets))
	for i := range tickets {
		data = append(data, toTicketResponse(&tickets[i]))
	}
	return PageResponse[TicketResponse]{Data: data, Page: page, Limit: limit, Total: total}
}
