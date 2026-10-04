package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestTableRollTriggersNothing — ADR 0121 §5: a die rolled at the table
// for fun is not an instruction from an effect (CR 706.1), so a
// "whenever you roll one or more dice" ability does not see it. Vexing
// Puzzlebox triggers on a card's roll (the control) and on no table
// roll, whatever the die.
func TestTableRollTriggersNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Vexing Puzzlebox", "Artifact", "7267ab16-2157-4b86-93ea-ca2c0fee064a", false)

	for _, die := range []string{game.TableDieD6, game.TableDieD20, game.TableDieCoin} {
		if _, err := g.RollTableDie(me.ID, die); err != nil {
			t.Fatalf("RollTableDie(%s): %v", die, err)
		}
	}
	if n := len(g.PendingTriggers); n != 0 {
		t.Fatalf("three table rolls queued %d triggers, want none", n)
	}
	if n := len(randomEvents(g, game.EventRollDie)) + len(randomEvents(g, game.EventFlipCoin)); n != 0 {
		t.Fatalf("table rolls emitted %d game roll events, want none", n)
	}

	// The control: a card's roll does trigger it.
	g.WithWriteLock(func() {
		if _, err := g.RollDiceForEffect(game.RandomDraw{Player: me.ID, Source: id}, 6, 1); err != nil {
			t.Fatal(err)
		}
	})
	if n := len(g.PendingTriggers); n != 1 {
		t.Fatalf("a card's roll queued %d triggers, want 1", n)
	}
}
