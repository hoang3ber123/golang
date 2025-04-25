package models

type Product struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Price       float64 `json:"price"`
	RelatedType string  `json:"related_type"`
	Image       string  `json:"image"`
}

func (*Product) GetTableName() string {
	return "products"
}
func (*Product) GetRelatedType() string {
	return "product"
}
