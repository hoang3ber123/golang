package serializers

type UserStatsResponse struct {
	TotalUsers      int64              `json:"total_users"`
	NewUsersInMonth int64              `json:"new_users_in_month"`
	UsersByMonth    []MonthlyUserCount `json:"users_by_month"`
}
type MonthlyUserCount struct {
	Name  int `json:"name"`
	Count int `json:"count"`
}
