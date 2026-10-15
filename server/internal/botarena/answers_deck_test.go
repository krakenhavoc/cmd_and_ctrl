package botarena_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/botarena"
)

// answers_deck_test.go is #2939's arena deck: BattleDeck with ten
// Basalt Gargoyles, so a table has a pump creature that takes several
// activations to live through a Lightning Bolt.

func TestAnswersBattleDeck(t *testing.T) {
	deck := botarena.AnswersBattleDeck([16]byte{})
	if len(deck) != len(botarena.BattleDeck([16]byte{})) {
		t.Errorf("%d cards, want BattleDeck's %d", len(deck), len(botarena.BattleDeck([16]byte{})))
	}
	n := 0
	for _, c := range deck {
		if c.Name == "Basalt Gargoyle" {
			n++
		}
	}
	if n != 10 {
		t.Errorf("%d Basalt Gargoyles, want 10", n)
	}
	if !botarena.IsSynthetic(botarena.AnswersDeckID) {
		t.Error("answers-battle needs no Scryfall dump")
	}
}
