package util

import (
	"fmt"
	"time"
)

// parseTime converts a []uint8 (byte slice) into a time.Time object.
// It expects the time to be in the "2006-01-02 15:04:05" format.
func ParseTime(data []uint8) (time.Time, error) {
	timeStr := string(data)

	// First try parsing with RFC3339
	parsedTime, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return parsedTime, nil
	}
	parsedTime, err = time.Parse("2006-01-02 15:04:05", timeStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing time: %v", err)
	}
	return parsedTime, nil
}

func ConvertStatus(status string) string {
	if status == "Approved" {
		return "appr"
	} else if status == "Pending" {
		return "pen"
	} else if status == "Accepted" {
		return "acc"
	} else if status == "Cancelled" {
		return "cancl"
	} else {
		return ""
	}
}
