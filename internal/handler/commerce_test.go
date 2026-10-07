package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"shop-api/internal/domain"
)

func TestCommerceResponsesUseFrontendJSONContract(t *testing.T) {
	shipping, err := json.Marshal(shippingMethodDTO(&domain.ShippingMethod{ID: 17, Name: "پست", Price: 125000, IsActive: true, SortOrder: 2}))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"id":17`, `"name":"پست"`, `"price":125000`, `"is_active":true`, `"sort_order":2`} {
		if !strings.Contains(string(shipping), field) {
			t.Errorf("shipping response %s is missing expected field %s", shipping, field)
		}
	}
	if strings.Contains(string(shipping), `"ID"`) || strings.Contains(string(shipping), `"Price"`) {
		t.Errorf("shipping response exposes Go field names: %s", shipping)
	}

	coupon, err := json.Marshal(couponDTO(&domain.Coupon{ID: 23, Code: "KICKS10", Kind: "percent", Value: 10, MinimumSubtotal: 0, UsedCount: 3, IsActive: true}))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"id":23`, `"code":"KICKS10"`, `"kind":"percent"`, `"value":10`, `"minimum_subtotal":0`, `"used_count":3`, `"is_active":true`} {
		if !strings.Contains(string(coupon), field) {
			t.Errorf("coupon response %s is missing expected field %s", coupon, field)
		}
	}
	if strings.Contains(string(coupon), `"ID"`) || strings.Contains(string(coupon), `"MinimumSubtotal"`) {
		t.Errorf("coupon response exposes Go field names: %s", coupon)
	}
}
