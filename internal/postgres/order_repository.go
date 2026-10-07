package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"shop-api/internal/domain"
)

type OrderRepository struct{ db DBTX }

func NewOrderRepository(db DBTX) *OrderRepository { return &OrderRepository{db: db} }

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	const insertOrder = `INSERT INTO orders (
		user_id,status,total_amount,items_amount,discount_amount,shipping_cost,coupon_id,coupon_code,shipping_method_id,shipping_method_name,tracking_code,
		shipping_receiver_name,shipping_receiver_phone,shipping_province,shipping_city,shipping_address_line,shipping_postal_code
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17) RETURNING id,created_at,updated_at`
	err := r.db.QueryRow(ctx, insertOrder,
		o.UserID, string(o.Status), o.TotalAmount, o.ItemsAmount, o.DiscountAmount, o.ShippingCost, o.CouponID, o.CouponCode, o.ShippingMethodID, o.ShippingMethodName, o.TrackingCode,
		o.ShippingReceiverName, o.ShippingReceiverPhone, o.ShippingProvince,
		o.ShippingCity, o.ShippingAddressLine, o.ShippingPostalCode,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert order: %w", err)
	}

	const insertItem = `INSERT INTO order_items (order_id, product_id, product_name, variant_id, variant_label, variant_name, quantity, unit_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	for i := range o.Items {
		item := &o.Items[i]
		item.OrderID = o.ID
		if err := r.db.QueryRow(ctx, insertItem, item.OrderID, item.ProductID, item.ProductName, item.VariantID, item.VariantLabel, item.VariantName, item.Quantity, item.UnitPrice).Scan(&item.ID); err != nil {
			return fmt.Errorf("insert order item: %w", err)
		}
	}
	return nil
}

const orderColumns = `o.id, o.user_id, o.status, o.total_amount, o.items_amount, o.discount_amount, o.shipping_cost, o.coupon_id, o.coupon_code, o.shipping_method_id, o.shipping_method_name, o.tracking_code,
	o.shipping_receiver_name, o.shipping_receiver_phone, o.shipping_province,
	o.shipping_city, o.shipping_address_line, o.shipping_postal_code,
	concat(u.first_name, ' ', u.last_name), u.phone, o.created_at, o.updated_at,
	p.authority, p.status, p.ref_id, p.paid_at`

const orderFrom = ` FROM orders o JOIN users u ON u.id = o.user_id LEFT JOIN payments p ON p.order_id = o.id `

func scanOrderWithPayment(row pgx.Row) (*domain.Order, error) {
	var (
		o                domain.Order
		status           string
		payAuthority     *string
		payStatus        *string
		payRefID         *string
		payPaidAt        *time.Time
		shippingMethodID pgtype.Int8
		couponID         pgtype.Int8
	)
	if err := row.Scan(
		&o.ID, &o.UserID, &status, &o.TotalAmount, &o.ItemsAmount, &o.DiscountAmount, &o.ShippingCost, &couponID, &o.CouponCode, &shippingMethodID, &o.ShippingMethodName, &o.TrackingCode,
		&o.ShippingReceiverName, &o.ShippingReceiverPhone, &o.ShippingProvince,
		&o.ShippingCity, &o.ShippingAddressLine, &o.ShippingPostalCode,
		&o.CustomerName, &o.CustomerPhone, &o.CreatedAt, &o.UpdatedAt,
		&payAuthority, &payStatus, &payRefID, &payPaidAt,
	); err != nil {
		return nil, err
	}
	o.Status = domain.OrderStatus(status)
	if shippingMethodID.Valid {
		id := shippingMethodID.Int64
		o.ShippingMethodID = &id
	}
	if couponID.Valid {
		id := couponID.Int64
		o.CouponID = &id
	}
	if payAuthority != nil {
		o.Payment = &domain.OrderPaymentInfo{Authority: *payAuthority, Status: domain.PaymentStatus(*payStatus), RefID: payRefID, PaidAt: payPaidAt}
	}
	return &o, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id int64) (*domain.Order, error) {
	return r.getByID(ctx, id, false)
}

func (r *OrderRepository) GetByIDForUpdate(ctx context.Context, id int64) (*domain.Order, error) {
	return r.getByID(ctx, id, true)
}

func (r *OrderRepository) GetByTrackingCode(ctx context.Context, code string) (*domain.Order, error) {
	o, err := scanOrderWithPayment(r.db.QueryRow(ctx, `SELECT `+orderColumns+orderFrom+`WHERE o.tracking_code=$1`, code))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order by tracking code: %w", err)
	}
	items, err := r.itemsByOrderIDs(ctx, []int64{o.ID})
	if err != nil {
		return nil, err
	}
	o.Items = items[o.ID]
	return o, nil
}

func (r *OrderRepository) getByID(ctx context.Context, id int64, lock bool) (*domain.Order, error) {
	query := `SELECT ` + orderColumns + orderFrom + `WHERE o.id = $1`
	if lock {
		query += ` FOR UPDATE OF o`
	}
	o, err := scanOrderWithPayment(r.db.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get order: %w", err)
	}
	items, err := r.itemsByOrderIDs(ctx, []int64{o.ID})
	if err != nil {
		return nil, err
	}
	o.Items = items[o.ID]
	return o, nil
}

func (r *OrderRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Order, int, error) {
	return r.list(ctx, userID, "", limit, offset)
}

func (r *OrderRepository) ListAll(ctx context.Context, status domain.OrderStatus, limit, offset int) ([]domain.Order, int, error) {
	return r.list(ctx, 0, string(status), limit, offset)
}

func (r *OrderRepository) list(ctx context.Context, userID int64, status string, limit, offset int) ([]domain.Order, int, error) {
	where := `($1 = 0 OR o.user_id = $1) AND ($2 = '' OR o.status = $2)`
	var total int
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM orders o WHERE `+where, userID, status).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count orders: %w", err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+orderColumns+orderFrom+`WHERE `+where+
		` ORDER BY o.id DESC LIMIT $3 OFFSET $4`, userID, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list orders: %w", err)
	}
	orders := make([]domain.Order, 0, limit)
	ids := make([]int64, 0, limit)
	for rows.Next() {
		o, err := scanOrderWithPayment(rows)
		if err != nil {
			rows.Close()
			return nil, 0, fmt.Errorf("scan order: %w", err)
		}
		orders = append(orders, *o)
		ids = append(ids, o.ID)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, 0, fmt.Errorf("iterate orders: %w", err)
	}
	rows.Close()
	if len(ids) == 0 {
		return orders, total, nil
	}
	items, err := r.itemsByOrderIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}
	for i := range orders {
		orders[i].Items = items[orders[i].ID]
	}
	return orders, total, nil
}

func (r *OrderRepository) TransitionStatus(ctx context.Context, id int64, from, to domain.OrderStatus) error {
	tag, err := r.db.Exec(ctx, `UPDATE orders SET status = $3, updated_at = now() WHERE id = $1 AND status = $2`, id, string(from), string(to))
	if err != nil {
		return fmt.Errorf("transition order status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := r.db.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM orders WHERE id = $1)`, id).Scan(&exists); err != nil {
			return fmt.Errorf("check order before status transition: %w", err)
		}
		if !exists {
			return domain.ErrNotFound
		}
		return domain.ErrOrderStatusConflict
	}
	return nil
}

func (r *OrderRepository) itemsByOrderIDs(ctx context.Context, orderIDs []int64) (map[int64][]domain.OrderItem, error) {
	rows, err := r.db.Query(ctx, `SELECT id, order_id, product_id, product_name, variant_id, variant_label, variant_name, quantity, unit_price
		FROM order_items WHERE order_id = ANY($1) ORDER BY id`, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()
	result := make(map[int64][]domain.OrderItem, len(orderIDs))
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.VariantID, &item.VariantLabel, &item.VariantName, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		result[item.OrderID] = append(result[item.OrderID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate order items: %w", err)
	}
	return result, nil
}
