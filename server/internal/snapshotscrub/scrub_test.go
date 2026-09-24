package snapshotscrub

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// A cut-down restore point with every kind of player data the real
// files carry. The IDs are shaped like the real ones.
const sample = `{
  "seq": 41,
  "snapshot": {
    "schema": 6,
    "id": "9b2f0c5e-1111-4a2b-8c3d-000000000001",
    "eventSeq": 1790000000000000001,
    "seats": [
      {"id": "7d3a1e2f-2222-4b3c-9d4e-000000000002", "name": "Luke", "displayName": "krakenhavoc",
       "discordId": "123456789012345678", "discordAvatarHash": "a1b2c3d4e5f6", "life": 40},
      {"id": "7d3a1e2f-3333-4b3c-9d4e-000000000003", "name": "Guest", "life": 37}
    ],
    "battlefield": {"kind": "battlefield", "cards": [
      {"instanceId": "c0ffee00-4444-4c5d-8e6f-000000000004", "name": "Sol Ring",
       "owner": "7d3a1e2f-2222-4b3c-9d4e-000000000002", "enteredBattlefieldAt": 1790000000123456789}
    ]},
    "commanderDamageBy": {"7d3a1e2f-3333-4b3c-9d4e-000000000003": 5},
    "events": [{"kind": "concede", "label": "game 9b2f0c5e-1111-4a2b-8c3d-000000000001 ended"}]
  }
}`

func TestScrubReplacesPlayerDataDeterministically(t *testing.T) {
	out, rep, err := Scrub([]byte(sample), Options{})
	if err != nil {
		t.Fatalf("Scrub: %v", err)
	}
	s := string(out)
	for _, gone := range []string{"Luke", "krakenhavoc", "123456789012345678", "a1b2c3d4e5f6",
		"7d3a1e2f-2222-4b3c-9d4e-000000000002", "7d3a1e2f-3333-4b3c-9d4e-000000000003",
		"9b2f0c5e-1111-4a2b-8c3d-000000000001", "discordId", "discordAvatarHash"} {
		if strings.Contains(s, gone) {
			t.Errorf("scrubbed output still contains %q", gone)
		}
	}
	for _, kept := range []string{`"Player 1"`, `"Player 2"`, SeatPlaceholderID(0), SeatPlaceholderID(1),
		"Sol Ring", "1790000000123456789", "1790000000000000001", `"seq": 41`} {
		if !strings.Contains(s, kept) {
			t.Errorf("scrubbed output lost %q", kept)
		}
	}
	if rep.Seats != 2 {
		t.Errorf("report counted %d seats, want 2", rep.Seats)
	}
	// The owner reference and the map key followed the seat.
	if strings.Count(s, SeatPlaceholderID(0)) < 2 || !strings.Contains(s, `"`+SeatPlaceholderID(1)+`": 5`) {
		t.Errorf("a player ID was not replaced everywhere:\n%s", s)
	}

	again, _, err := Scrub([]byte(sample), Options{})
	if err != nil || !bytes.Equal(out, again) {
		t.Error("Scrub is not deterministic")
	}
	// Scrubbing a scrubbed file changes nothing but is still accepted.
	if _, _, err := Scrub(out, Options{}); err != nil {
		t.Errorf("re-scrubbing a scrubbed file: %v", err)
	}
}

func TestScrubRefusesLeftoverSnowflakesAndEmails(t *testing.T) {
	for name, extra := range map[string]string{
		"snowflake": `"a note from 987654321098765432"`,
		"email":     `"mail someone@example.com"`,
	} {
		raw := strings.Replace(sample, `"events": [`, `"notes": [`+extra+`], "events": [`, 1)
		if _, _, err := Scrub([]byte(raw), Options{}); !errors.Is(err, ErrRefused) {
			t.Errorf("%s: Scrub returned %v, want ErrRefused", name, err)
		}
	}
}

func TestScrubRefusesAnEmbeddedNameUnlessReviewed(t *testing.T) {
	raw := strings.Replace(sample, `"name": "Sol Ring"`, `"name": "Luke's Sol Ring"`, 1)
	if _, _, err := Scrub([]byte(raw), Options{}); !errors.Is(err, ErrRefused) {
		t.Fatalf("Scrub returned %v, want ErrRefused", err)
	}
	_, rep, err := Scrub([]byte(raw), Options{AllowNameSubstrings: true})
	if err != nil || len(rep.Warnings) == 0 {
		t.Errorf("reviewed mode: err=%v warnings=%v, want a warning and no error", err, rep.Warnings)
	}
}

func TestScrubRejectsSomethingThatIsNotARestorePoint(t *testing.T) {
	if _, _, err := Scrub([]byte(`{"seq": 1, "snapshot": {"id": "x"}}`), Options{}); err == nil {
		t.Error("a file with no schema was accepted")
	}
}
