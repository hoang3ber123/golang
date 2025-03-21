package services

import (
	"time"
)

// Hàm tính số lượng time_unit giữa startDay và endDay
func CountTimeUnits(startDay, endDay, timeUnit string) int {
	// Định dạng ngày
	layout := "2006-01-02"

	// Chuyển đổi string thành time.Time
	start, err := time.Parse(layout, startDay)
	if err != nil {
		return 0
	}
	end, err := time.Parse(layout, endDay)
	if err != nil {
		return 0
	}

	// Tính toán theo từng loại timeUnit
	switch timeUnit {
	case "day":
		return int(end.Sub(start).Hours()/24) + 1
	case "month":
		yearsDiff := end.Year() - start.Year()
		monthsDiff := yearsDiff*12 + int(end.Month()) - int(start.Month())
		if end.Day() > start.Day() {
			monthsDiff++ // Nếu ngày của end lớn hơn ngày của start, tăng thêm 1 tháng
		}
		return monthsDiff
	case "year":
		yearsDiff := end.Year() - start.Year()
		if end.YearDay() > start.YearDay() {
			yearsDiff++ // Nếu ngày trong năm của end lớn hơn start, tăng thêm 1 năm
		}
		return yearsDiff
	default:
		return 0
	}
}

// parseDate phân tích chuỗi ngày theo định dạng "YYYY-MM-DD" và trả về đối tượng time.Time
func parseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// endOfMonth trả về ngày cuối cùng của tháng từ chuỗi ngày đầu vào
func EndOfMonth(dateStr string) (string, error) {
	t, err := parseDate(dateStr)
	if err != nil {
		return "", err
	}
	// Chuyển đến ngày 1 của tháng tiếp theo
	nextMonth := t.AddDate(0, 1, 0)
	firstOfNextMonth := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, t.Location())
	// Trừ đi một ngày để có ngày cuối của tháng hiện tại
	endOfMonth := firstOfNextMonth.AddDate(0, 0, -1)
	return endOfMonth.Format("2006-01-02"), nil
}

// endOfYear trả về ngày cuối cùng của năm từ chuỗi ngày đầu vào
func EndOfYear(dateStr string) (string, error) {
	t, err := parseDate(dateStr)
	if err != nil {
		return "", err
	}
	// Chuyển đến ngày 1 tháng 1 của năm tiếp theo
	nextYear := t.AddDate(1, 0, 0)
	firstOfNextYear := time.Date(nextYear.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	// Trừ đi một ngày để có ngày cuối của năm hiện tại
	endOfYear := firstOfNextYear.AddDate(0, 0, -1)
	return endOfYear.Format("2006-01-02"), nil
}

// startOfMonth trả về ngày đầu tiên của tháng từ chuỗi ngày đầu vào
func StartOfMonth(dateStr string) (string, error) {
	t, err := parseDate(dateStr)
	if err != nil {
		return "", err
	}
	// Tạo ngày 1 của tháng hiện tại
	startOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	return startOfMonth.Format("2006-01-02"), nil
}

// startOfYear trả về ngày đầu tiên của năm từ chuỗi ngày đầu vào
func StartOfYear(dateStr string) (string, error) {
	t, err := parseDate(dateStr)
	if err != nil {
		return "", err
	}
	// Tạo ngày 1 tháng 1 của năm hiện tại
	startOfYear := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	return startOfYear.Format("2006-01-02"), nil
}
