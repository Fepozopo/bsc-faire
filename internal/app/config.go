package app

import "time"

type Order struct {
	ID                       string    `json:"id"`
	DisplayID                string    `json:"display_id"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
	State                    string    `json:"state"`
	IsFreeShipping           bool      `json:"is_free_shipping"`
	FreeShippingReason       string    `json:"free_shipping_reason"`
	FaireCoveredShippingCost struct {
		AmountMinor int    `json:"amount_minor"`
		Currency    string `json:"currency"`
	} `json:"faire_covered_shipping_cost"`
	Items []struct {
		ID             string    `json:"id"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		OrderID        string    `json:"order_id"`
		ProductID      string    `json:"product_id"`
		VariantID      string    `json:"variant_id"`
		Quantity       int       `json:"quantity"`
		Sku            string    `json:"sku"`
		PriceCents     int       `json:"price_cents"`
		ProductName    string    `json:"product_name"`
		VariantName    string    `json:"variant_name"`
		IncludesTester bool      `json:"includes_tester"`
		Discounts      []any     `json:"discounts"`
	} `json:"items"`
	Shipments []struct {
		ID             string    `json:"id"`
		OrderID        string    `json:"order_id"`
		MakerCostCents int       `json:"maker_cost_cents"`
		Carrier        string    `json:"carrier"`
		TrackingCode   string    `json:"tracking_code"`
		CreatedAt      time.Time `json:"created_at"`
		UpdatedAt      time.Time `json:"updated_at"`
		ShippingType   string    `json:"shipping_type"`
	} `json:"shipments"`
	Address struct {
		Name        string `json:"name"`
		Address1    string `json:"address1"`
		Address2    string `json:"address2"`
		PostalCode  string `json:"postal_code"`
		City        string `json:"city"`
		State       string `json:"state"`
		StateCode   string `json:"state_code"`
		PhoneNumber string `json:"phone_number"`
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
		CompanyName string `json:"company_name"`
	} `json:"address"`
	RetailerID  string    `json:"retailer_id"`
	ShipAfter   time.Time `json:"ship_after"`
	PayoutCosts struct {
		PayoutFeeCents              int `json:"payout_fee_cents"`
		PayoutFeeBps                int `json:"payout_fee_bps"`
		CommissionCents             int `json:"commission_cents"`
		CommissionBps               int `json:"commission_bps"`
		SubtotalAfterBrandDiscounts struct {
			AmountMinor int    `json:"amount_minor"`
			Currency    string `json:"currency"`
		} `json:"subtotal_after_brand_discounts"`
		TotalBrandDiscounts struct {
			AmountMinor int    `json:"amount_minor"`
			Currency    string `json:"currency"`
		} `json:"total_brand_discounts"`
	} `json:"payout_costs"`
	Source             string    `json:"source"`
	PaymentInitiatedAt time.Time `json:"payment_initiated_at"`
	OriginalOrderID    string    `json:"original_order_id"`
	BrandDiscounts     []struct {
		ID                   string  `json:"id"`
		Code                 string  `json:"code"`
		IncludesFreeShipping bool    `json:"includes_free_shipping"`
		DiscountPercentage   float64 `json:"discount_percentage"`
		DiscountType         string  `json:"discount_type"`
	} `json:"brand_discounts"`
	EstimatedPayoutAt time.Time `json:"estimated_payout_at"`
}

type Orders struct {
	Page   int     `json:"page"`
	Limit  int     `json:"limit"`
	Cursor string  `json:"cursor"`
	Orders []Order `json:"orders"`
}
