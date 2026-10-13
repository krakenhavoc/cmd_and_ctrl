package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// conditional_exile_test.go — #3192. The enumerator (and so the bot)
// offers a conditional exile play only while its condition holds. It
// reads the same CastPermissionForLocked the cast path does, so the
// move disappears while the condition is false and comes back when it
// holds, and the cast the move would make is accepted exactly then.
func TestEnumeratorOffersAConditionalExilePlayOnlyWhileItHolds(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 4, "Mountain")

	bolt := game.NewCard("Held Bolt", seat.ID)
	bolt.TypeLine = "Instant"
	bolt.ManaCost = "{R}"
	seat.Library.PushTop(bolt)
	g.WithWriteLock(func() {
		_, err := g.ExileTopWithPermissionForEffect(seat.ID, seat.ID, 1, game.CastPermission{
			Duration:  game.WhileInZoneDuration(),
			Condition: &game.PermissionCondition{Key: game.CondControlSubtype, Subtype: "Kavu"},
		})
		if err != nil {
			t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
		}
	})

	if got := castMovesFor(legal.EnumerateFor(g, seat.ID), bolt.InstanceID); len(got) != 0 {
		t.Fatalf("a dormant exile play was offered: %v", labels(got))
	}

	kavu := creature("Kavu", "{2}{R}", 2, 2)
	kavu.TypeLine = "Creature — Kavu"
	battlefieldCard(g, seat, kavu)
	moves := legal.EnumerateFor(g, seat.ID)
	got := castMovesFor(moves, bolt.InstanceID)
	if len(got) == 0 {
		t.Fatalf("the condition holds and no cast was offered: %v", labels(moves))
	}
	dispatchAll(t, g, seat.ID, got)
}
