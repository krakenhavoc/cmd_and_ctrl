package game

import "testing"

// TestFlashbackInTextFindsTheKeywordLine pins the reader against Oracle
// text shapes: a mana cost, a non-mana cost, reminder text, and a card
// that only names flashback.
func TestFlashbackInTextFindsTheKeywordLine(t *testing.T) {
	for _, tc := range []struct {
		text string
		want bool
	}{
		{"Deal 2 damage to any target.\nFlashback {4}{R} (You may cast this card from your graveyard for its flashback cost. Then exile it.)", true},
		{"Return three target creature cards.\nFlashback—Sacrifice three creatures.", true},
		{"Flashback—{1}{G}, Pay 3 life.", true},
		{"Each instant and sorcery card in your graveyard has flashback. Its flashback cost is equal to its mana cost.", false},
		{"Flashback costs you pay cost {1} less.", false},
		{"Draw a card.", false},
		{"", false},
	} {
		if got := FlashbackInText(tc.text); got != tc.want {
			t.Errorf("FlashbackInText(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}

func TestHasFlashbackReadsTheStamp(t *testing.T) {
	if HasFlashback(Card{}) {
		t.Error("an unread card with no catalog entry has no flashback")
	}
	if !HasFlashback(Card{FlashbackText: PrintedYes}) {
		t.Error("a card whose text prints flashback has it")
	}
	if HasFlashback(Card{FlashbackText: PrintedYes, FaceDown: true}) {
		t.Error("a face-down card has no text")
	}
}
