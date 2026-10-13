package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// animate_dead_test.go — ADR 0154 §9 (#1780). Animate Dead's enchant
// clause is the graveyard target clause Reanimate uses, so the
// enumerator offers one cast per creature card in any graveyard and no
// other, and dispatchAll proves CastSpell accepts every one.

const oracleAnimateDead = "c0d8fef4-65f4-4769-982d-b397d2b7e977"

func TestEnumeratorOffersAnimateDeadOnEachGraveyardCreature(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Swamp")

	mine := graveyardCard(seat, creature("My Zombie", "{1}{B}", 2, 2))
	theirs := graveyardCard(other, creature("Their Beast", "{3}{G}", 4, 4))
	sorcery := graveyardCard(other, game.Card{Name: "Their Sorcery", TypeLine: "Sorcery", ManaCost: "{1}{R}"})
	onField := battlefieldCard(g, other, creature("Live Beast", "{1}{G}", 2, 2))

	aura := handCard(seat, game.Card{
		Name: "Animate Dead", TypeLine: "Enchantment — Aura", OracleID: oracleAnimateDead,
		ManaCost: "{1}{B}", Colors: []string{"B"},
	})

	moves := legal.EnumerateFor(g, seat.ID)
	casts := castPayloadsOf(t, moves, aura)
	seen := map[string]int{}
	for _, c := range casts {
		if len(c.Targets) != 1 {
			t.Errorf("a cast names %d targets, want 1: %+v", len(c.Targets), c)
			continue
		}
		seen[c.Targets[0].ID]++
	}
	if seen[mine.String()] != 1 || seen[theirs.String()] != 1 {
		t.Errorf("want one cast onto each graveyard creature card, got %+v: %v", seen, labels(moves))
	}
	if seen[sorcery.String()] != 0 || seen[onField.String()] != 0 {
		t.Errorf("offered a target the clause does not allow: %+v", seen)
	}
	if len(casts) != 2 {
		t.Errorf("%d casts offered, want 2", len(casts))
	}
	dispatchAll(t, g, seat.ID, moves)
}
