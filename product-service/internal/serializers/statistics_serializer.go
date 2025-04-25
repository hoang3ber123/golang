package serializers

type ProductStatsResponse struct {
	TotalProducts      int64                  `json:"total_products"`
	NewProductsInMonth int64                  `json:"new_products_in_month"`
	ProductsByMonth    []MonthlyProductCount  `json:"products_by_month"`
	ProductsByCategory []CategoryProductCount `json:"products_by_category"`
}

type MonthlyProductCount struct {
	Name  int `json:"name"`
	Count int `json:"count"`
}

type CategoryProductCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}
