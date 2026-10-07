package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id int64) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	List(ctx context.Context, search string, limit, offset int) ([]User, int, error)
	// UpdateAccess must be called within a transaction. It locks active admins
	// and prevents removing the final active admin.
	UpdateAccess(ctx context.Context, id int64, role Role, active bool) error
}

type Category struct {
	ID             int64
	Name           string
	Slug           string
	Description    string
	SEOTitle       string
	SEODescription string
	HomeTitle      string
	HomeImageURL   string
	ShowOnHome     bool
	ParentID       *int64
	ParentName     *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CategoryRepository interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, id int64) (*Category, error)
	GetBySlug(ctx context.Context, slug string) (*Category, error)
	List(ctx context.Context) ([]Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, id int64) error
	ValidateParent(ctx context.Context, id int64, parentID *int64) error
}

type Brand struct {
	ID             int64
	Name           string
	Slug           string
	Description    string
	SEOTitle       string
	SEODescription string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BrandRepository interface {
	Create(ctx context.Context, brand *Brand) error
	GetByID(ctx context.Context, id int64) (*Brand, error)
	GetBySlug(ctx context.Context, slug string) (*Brand, error)
	List(ctx context.Context) ([]Brand, error)
	Update(ctx context.Context, brand *Brand) error
	Delete(ctx context.Context, id int64) error
}

type Banner struct {
	ID              int64
	Title           string
	Subtitle        string
	DesktopImageURL string
	MobileImageURL  string
	LinkURL         string
	ButtonLabel     string
	SortOrder       int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type BannerRepository interface {
	ListActive(ctx context.Context) ([]Banner, error)
	ListAll(ctx context.Context) ([]Banner, error)
	Create(ctx context.Context, banner *Banner) error
	Update(ctx context.Context, banner *Banner) error
	Delete(ctx context.Context, id int64) error
}

type ShippingMethod struct {
	ID          int64
	Name        string
	Description string
	Province    string
	City        string
	Price       int64
	IsActive    bool
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ShippingMethodRepository interface {
	ListAvailable(ctx context.Context, province, city string) ([]ShippingMethod, error)
	ListAll(ctx context.Context) ([]ShippingMethod, error)
	GetByID(ctx context.Context, id int64) (*ShippingMethod, error)
	Create(ctx context.Context, method *ShippingMethod) error
	Update(ctx context.Context, method *ShippingMethod) error
	Delete(ctx context.Context, id int64) error
}

type Coupon struct {
	ID              int64
	Code            string
	Kind            string
	Value           int64
	MinimumSubtotal int64
	MaximumDiscount *int64
	UsageLimit      *int64
	UsedCount       int64
	StartsAt        *time.Time
	EndsAt          *time.Time
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CouponRepository interface {
	GetByCode(ctx context.Context, code string) (*Coupon, error)
	GetByCodeForUpdate(ctx context.Context, code string) (*Coupon, error)
	ListAll(ctx context.Context) ([]Coupon, error)
	Create(ctx context.Context, coupon *Coupon) error
	Update(ctx context.Context, coupon *Coupon) error
	Delete(ctx context.Context, id int64) error
	ReserveUse(ctx context.Context, id int64, now time.Time) error
	ReleaseUse(ctx context.Context, id int64) error
}

// OTPCode is the currently-active one-time-password state for a phone
// number. There is at most one row per phone: requesting a new code
// replaces it.
type OTPCode struct {
	Phone      string
	CodeHash   string
	Attempts   int
	ExpiresAt  time.Time
	LastSentAt time.Time
}

type OTPRepository interface {
	// Upsert replaces any existing code for the phone and resets attempts to 0.
	Upsert(ctx context.Context, phone, codeHash string, expiresAt time.Time) error
	Get(ctx context.Context, phone string) (*OTPCode, error)
	IncrementAttempts(ctx context.Context, phone string) error
	Delete(ctx context.Context, phone string) error
}

// AddressRepository methods are scoped by userID so a user can never read,
// edit or delete another user's address, and a missing/foreign id reads the
// same as "not found" instead of leaking whether it exists.
type AddressRepository interface {
	Create(ctx context.Context, a *Address) error
	GetByID(ctx context.Context, userID, id int64) (*Address, error)
	ListByUserID(ctx context.Context, userID int64) ([]Address, error)
	Update(ctx context.Context, a *Address) error
	Delete(ctx context.Context, userID, id int64) error
	// SetDefault marks exactly one address (id) as default for the user and
	// unmarks the rest, atomically.
	SetDefault(ctx context.Context, userID, id int64) error
}

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	GetByID(ctx context.Context, id int64) (*Product, error)
	GetBySlug(ctx context.Context, slug string) (*Product, error)
	List(ctx context.Context, search string, categoryID, brandID *int64, limit, offset int) ([]Product, int, error)
	ListAdmin(ctx context.Context, search string, categoryID, brandID *int64, limit, offset int) ([]Product, int, error)
	Update(ctx context.Context, p *Product) error
	SetActive(ctx context.Context, id int64, active bool) error
	Delete(ctx context.Context, id int64) error

	// GetByIDsForUpdate loads products and locks their rows until the
	// surrounding transaction ends. Rows are returned ordered by id.
	GetByIDsForUpdate(ctx context.Context, ids []int64) ([]Product, error)
	// DecreaseStock atomically subtracts qty; returns ErrInsufficientStock if not enough is left.
	DecreaseStock(ctx context.Context, id int64, qty int) error
	DecreaseVariantStock(ctx context.Context, productID int64, variantID string, qty int) error
	// IncreaseStock adds qty back (used when a payment is cancelled).
	IncreaseStock(ctx context.Context, id int64, qty int) error
	IncreaseVariantStock(ctx context.Context, productID int64, variantID string, qty int) error
}

type ProductComment struct {
	ID          int64
	ProductID   int64
	ProductName string
	ProductSlug string
	UserID      int64
	AuthorName  string
	Body        string
	IsApproved  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CommentRepository interface {
	Create(ctx context.Context, comment *ProductComment) error
	ListApprovedByProduct(ctx context.Context, productID int64) ([]ProductComment, error)
	ListAll(ctx context.Context, approved *bool, limit, offset int) ([]ProductComment, int, error)
	SetApproval(ctx context.Context, id int64, approved bool) error
}

type BlogPost struct {
	ID             int64
	Title          string
	Slug           string
	Summary        string
	Content        string
	CoverImageURL  string
	SEOTitle       string
	SEODescription string
	IsPublished    bool
	PublishedAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type BlogRepository interface {
	Create(ctx context.Context, post *BlogPost) error
	GetByID(ctx context.Context, id int64) (*BlogPost, error)
	GetBySlug(ctx context.Context, slug string) (*BlogPost, error)
	ListPublished(ctx context.Context, search string, limit, offset int) ([]BlogPost, int, error)
	ListAdmin(ctx context.Context, search string, limit, offset int) ([]BlogPost, int, error)
	Update(ctx context.Context, post *BlogPost) error
	Delete(ctx context.Context, id int64) error
}

type TicketStatus string

const (
	TicketStatusOpen     TicketStatus = "open"
	TicketStatusAnswered TicketStatus = "answered"
	TicketStatusClosed   TicketStatus = "closed"
)

type SupportTicket struct {
	ID          int64
	UserID      int64
	UserName    string
	UserPhone   string
	Subject     string
	Status      TicketStatus
	LastMessage string
	Messages    []TicketMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TicketMessage struct {
	ID         int64
	TicketID   int64
	SenderID   int64
	SenderName string
	SenderRole Role
	Body       string
	CreatedAt  time.Time
}

type TicketRepository interface {
	Create(ctx context.Context, ticket *SupportTicket) error
	GetByID(ctx context.Context, id int64) (*SupportTicket, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*SupportTicket, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]SupportTicket, int, error)
	ListAll(ctx context.Context, status TicketStatus, limit, offset int) ([]SupportTicket, int, error)
	ListMessages(ctx context.Context, ticketID int64) ([]TicketMessage, error)
	AddMessage(ctx context.Context, message *TicketMessage) error
	SetStatus(ctx context.Context, id int64, status TicketStatus) error
}

type OrderRepository interface {
	// Create inserts the order and its items. Call it inside a transaction.
	Create(ctx context.Context, o *Order) error
	GetByID(ctx context.Context, id int64) (*Order, error)
	GetByTrackingCode(ctx context.Context, code string) (*Order, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*Order, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]Order, int, error)
	ListAll(ctx context.Context, status OrderStatus, limit, offset int) ([]Order, int, error)
	TransitionStatus(ctx context.Context, id int64, from, to OrderStatus) error
}

type DashboardStats struct {
	Products      int64
	Categories    int64
	Users         int64
	Orders        int64
	PendingOrders int64
	PaidRevenue   int64
}

type DashboardRepository interface {
	GetDashboardStats(ctx context.Context) (DashboardStats, error)
}

type SiteContent struct {
	ContactPhone   string
	ContactEmail   string
	ContactAddress string
	ContactHours   string
	InstagramURL   string
	AboutContent   string
	TermsContent   string
	UpdatedAt      time.Time
}

type SiteContentRepository interface {
	GetSiteContent(ctx context.Context) (*SiteContent, error)
	UpdateSiteContent(ctx context.Context, content *SiteContent) error
}

type AnalyticsVisitor struct {
	VisitorID string
	Path      string
	LastSeen  time.Time
}

type AnalyticsOverview struct {
	ActiveOnline   int64
	VisitorsToday  int64
	UniqueVisitors int64
	PageViewsToday int64
	ActiveVisitors []AnalyticsVisitor
}

type AnalyticsRepository interface {
	Heartbeat(ctx context.Context, visitorID, path string) error
	RecordPageView(ctx context.Context, viewID, visitorID, path string) error
	GetAnalyticsOverview(ctx context.Context) (AnalyticsOverview, error)
}

type PaymentRepository interface {
	Create(ctx context.Context, p *Payment) error
	GetByAuthority(ctx context.Context, authority string) (*Payment, error)
	GetByOrderID(ctx context.Context, orderID int64) (*Payment, error)
	MarkPaid(ctx context.Context, authority, refID string) error
	MarkCancelled(ctx context.Context, authority string) error
}

// Repositories groups repositories that share the same database handle
// (a connection pool, or a single transaction).
type Repositories struct {
	Users      UserRepository
	Categories CategoryRepository
	Brands     BrandRepository
	Products   ProductRepository
	Orders     OrderRepository
	Payments   PaymentRepository
	Tickets    TicketRepository
	Shipping   ShippingMethodRepository
	Coupons    CouponRepository
}

// TxManager runs fn inside one database transaction. If fn returns an error
// (or panics) everything is rolled back, otherwise it is committed.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(repos Repositories) error) error
}
