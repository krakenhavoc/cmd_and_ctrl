package game

import "testing"

// protection_card_name_test.go — ProtectionQualityCardName, "protection from
// the card named X" (Runed Halo).

func TestProtectionFromACardNameParsesAndMatchesBySourceName(t *testing.T) {
	tok := ProtectionFromCardName("Lightning Bolt")
	if tok != "protection from the card named Lightning Bolt" {
		t.Fatalf("token %q", tok)
	}
	q, ok := ParseProtectionQuality(tok)
	if !ok || q.Kind != ProtectionQualityCardName || q.Value != "Lightning Bolt" {
		t.Fatalf("parsed %+v, %v", q, ok)
	}
	if q.Token() != tok {
		t.Errorf("round trip %q, want %q", q.Token(), tok)
	}
	if !q.Matches(&Characteristic{Name: "lightning bolt"}) {
		t.Error("a source with that name (any case) is not stopped")
	}
	if q.Matches(&Characteristic{Name: "Shock"}) || q.Matches(nil) {
		t.Error("another source is stopped")
	}
}

func TestProtectionFromAnEmptyCardNameIsNoQuality(t *testing.T) {
	if ProtectionFromCardName("  ") != "" {
		t.Error("an empty name made a token")
	}
	if _, ok := ParseProtectionQuality("protection from the card named "); ok {
		t.Error("an empty name parsed")
	}
}
