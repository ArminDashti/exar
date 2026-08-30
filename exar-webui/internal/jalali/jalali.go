package jalali

import "fmt"

var gdm = []int{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334}
var jMonthLen = []int{31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29}

// ToJalali converts a Gregorian date to Jalali year/month/day.
func ToJalali(gy, gm, gd int) (jy, jm, jd int) {
	gy2 := gy - 1600
	gm2 := gm - 1
	gd2 := gd - 1

	gDayNo := 365*gy2 + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400
	gDayNo += gdm[gm2] + gd2
	if gm > 2 && ((gy%4 == 0 && gy%100 != 0) || (gy%400 == 0)) {
		gDayNo++
	}

	jDayNo := gDayNo - 79
	jNp := jDayNo / 12053
	jDayNo %= 12053
	jy = 979 + 33*jNp + 4*(jDayNo/1461)
	jDayNo %= 1461

	if jDayNo >= 366 {
		jy += (jDayNo - 1) / 365
		jDayNo = (jDayNo - 1) % 365
	}

	var i int
	for i = 0; i < 11 && jDayNo >= jMonthLen[i]; i++ {
		jDayNo -= jMonthLen[i]
	}
	jm = i + 1
	jd = jDayNo + 1
	return jy, jm, jd
}

// MonthKeyFromGregorian returns "YYYY/MM" Jalali month key from Gregorian YYYY-MM-DD.
func MonthKeyFromGregorian(date string) (string, error) {
	var gy, gm, gd int
	if _, err := fmt.Sscanf(date, "%d-%d-%d", &gy, &gm, &gd); err != nil {
		return "", fmt.Errorf("invalid date %q", date)
	}
	jy, jm, _ := ToJalali(gy, gm, gd)
	return fmt.Sprintf("%04d/%02d", jy, jm), nil
}
