package main

import (
	"testing"
	"time"
)

func TestCronAtUsesUTCCalendarFields(t *testing.T) {
	runAt := time.Date(2026, time.August, 4, 17, 35, 0, 0, time.UTC)
	if got, want := cronAt(runAt), "35 17 4 8 *"; got != want {
		t.Fatalf("cronAt() = %q, want %q", got, want)
	}
}

func TestIdempotencyKeyIsStable(t *testing.T) {
	runAt := time.Date(2026, time.August, 4, 17, 35, 0, 0, time.UTC)
	first := idempotencyKey("care-plan-1842", runAt)
	second := idempotencyKey("care-plan-1842", runAt)
	if first != second {
		t.Fatalf("idempotency key changed: %q != %q", first, second)
	}
}
