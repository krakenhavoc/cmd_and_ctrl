package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
)

// answers_deck_test.go is #2939's arena deck: BattleDeck with ten
// Basalt Gargoyles, so a table has a pump creature that takes several
// activations to live through a Lightning Bolt, and eight Selfless
// Spirits.

func TestAnswersBattleDeck(t *testing.T) {
	deck := botarena.AnswersBattleDeck([16]byte{})
	if len(deck) != len(botarena.BattleDeck([16]byte{})) {
		t.Errorf("%d cards, want BattleDeck's %d", len(deck), len(botarena.BattleDeck([16]byte{})))
	}
	n := map[string]int{}
	for _, c := range deck {
		n[c.Name]++
	}
	if n["Basalt Gargoyle"] != 10 || n["Selfless Spirit"] != 8 || n["Plains"] != 8 {
		t.Errorf("%d Basalt Gargoyles, %d Selfless Spirits, %d Plains, want 10, 8 and 8", n["Basalt Gargoyle"], n["Selfless Spirit"], n["Plains"])
	}
	if !botarena.IsSynthetic(botarena.AnswersDeckID) {
		t.Error("answers-battle needs no Scryfall dump")
	}
}
