package jalali

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	ptime "github.com/yaa110/go-persian-calendar"
)

// ToGregorian parses a Persian calendar date string (YYYY-MM-DD) and returns
// the equivalent Gregorian date formatted as YYYY-MM-DD.
func ToGregorian(persianDate string) (string, error) {
	parts := strings.Split(strings.TrimSpace(persianDate), "-")
	if len(parts) != 3 {
		return "", fmt.Errorf("date must be in Persian calendar format YYYY-MM-DD (e.g. 1405-06-10)")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("invalid year in date: %w", err)
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil || month < 1 || month > 12 {
		return "", fmt.Errorf("invalid month in date: must be 1-12")
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil || day < 1 || day > 31 {
		return "", fmt.Errorf("invalid day in date: must be 1-31")
	}

	pt := ptime.Date(year, ptime.Month(month), day, 0, 0, 0, 0, ptime.Iran())
	if pt.IsZero() {
		return "", fmt.Errorf("invalid Persian date: %s", persianDate)
	}

	gregorian := pt.Time()
	return gregorian.Format(time.DateOnly), nil
}
