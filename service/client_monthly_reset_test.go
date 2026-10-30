package service

import (
	"testing"
	"time"
)

func TestCalculateMonthlyReset(t *testing.T) {
	loc := time.UTC

	// Test 1: Created on Oct 30 -> resets on Nov 30 (31 days)
	oct30 := time.Date(2026, 10, 30, 15, 0, 0, 0, loc)
	days, nextReset := calculateMonthlyReset(oct30)
	if days != 31 {
		t.Fatalf("Oct 30 should have 31 reset days to Nov 30, got %d", days)
	}
	nextDate := time.Unix(nextReset, 0).In(loc)
	if nextDate.Month() != 11 || nextDate.Day() != 30 {
		t.Fatalf("Oct 30 should reset on Nov 30, got %v", nextDate)
	}

	// Test 2: Created on Jan 31 -> resets by 30 days
	jan31 := time.Date(2026, 1, 31, 12, 0, 0, 0, loc)
	days, nextReset = calculateMonthlyReset(jan31)
	if days != 30 {
		t.Fatalf("Jan 31 should have 30 reset days, got %d", days)
	}
	expectedTarget := jan31.AddDate(0, 0, 30).Unix()
	if nextReset != expectedTarget {
		t.Fatalf("Jan 31 target timestamp mismatch: got %d want %d", nextReset, expectedTarget)
	}

	// Test 3: Created on Jan 15 -> resets on Feb 15 (31 days)
	jan15 := time.Date(2026, 1, 15, 12, 0, 0, 0, loc)
	days, nextReset = calculateMonthlyReset(jan15)
	if days != 31 {
		t.Fatalf("Jan 15 should have 31 reset days, got %d", days)
	}
	nextDate = time.Unix(nextReset, 0).In(loc)
	if nextDate.Month() != 2 || nextDate.Day() != 15 {
		t.Fatalf("Jan 15 should reset on Feb 15, got %v", nextDate)
	}

	// Test 4: Year rollover Dec 20 -> Jan 20
	dec20 := time.Date(2026, 12, 20, 12, 0, 0, 0, loc)
	days, nextReset = calculateMonthlyReset(dec20)
	nextDate = time.Unix(nextReset, 0).In(loc)
	if nextDate.Year() != 2027 || nextDate.Month() != 1 || nextDate.Day() != 20 {
		t.Fatalf("Dec 20 should reset on Jan 20 next year, got %v", nextDate)
	}
}
