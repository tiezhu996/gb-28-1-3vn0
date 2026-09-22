package util

import "strconv"

func fmtSscan(s string, v *float64) (int, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	*v = f
	return 1, nil
}
