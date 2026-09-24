package ws

import "testing"

// TestInsufficientManaMessageNamesTheAnnouncement: #1296 — a refused
// activation says it was an activation. The cast wording is unchanged
// (the client's override toast keys off the code and card_id, never
// the sentence).
func TestInsufficientManaMessageNamesTheAnnouncement(t *testing.T) {
	cases := map[string]string{
		"cast_spell":       "insufficient mana to cast",
		"activate_ability": "insufficient mana to activate that ability",
	}
	for typ, want := range cases {
		if got := insufficientManaMessage(typ); got != want {
			t.Errorf("%s: %q, want %q", typ, got, want)
		}
	}
}
