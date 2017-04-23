package main

import (
	"time"
)

type RequestAuditDuration int
const (
	This_Week RequestAuditDuration = iota
	This_Month
	Last_Month
	Last_Three_Month
	Half_Year // This is the maximum duration GSuite can pull off: https://developers.google.com/admin-sdk/reports/v1/reference/activities/list?authuser=1
)

func (r RequestAuditDuration) ModifyDate(t time.Time) time.Time {
	switch r {
	case This_Week:
		for t.Weekday() != time.Monday {
			t = t.AddDate(0, 0, -1)
		}
	case This_Month:
		t = t.AddDate(0, 0, -(t.Day() - 1))
	case Last_Month:
		t = t.AddDate(0, -1, -(t.Day() - 1))
	case Last_Three_Month:
		t = t.AddDate(0, -3, -(t.Day() - 1))
	case Half_Year:
		t = t.AddDate(0, -6, -(t.Day() - 1))
	}
	return t
}