package aiseat

import "testing"

func TestCoinCallUsesUUIDLastBit(t *testing.T) {
	if got := CoinCall("00000000-0000-4000-8000-000000000000"); got != "heads" {
		t.Fatalf("even UUID call = %q", got)
	}
	if got := CoinCall("00000000-0000-4000-8000-000000000001"); got != "tails" {
		t.Fatalf("odd UUID call = %q", got)
	}
	if got := CoinCall("not-a-uuid"); got != "heads" {
		t.Fatalf("invalid UUID fallback = %q", got)
	}
}
