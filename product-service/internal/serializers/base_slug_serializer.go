package serializers

// BaseResponseSerializer chứa thông tin chung
type BaseSlugResponseSerializer struct {
	BaseResponseSerializer
	Title string `json:"title"`
	Slug  string `json:"slug"`
}
