package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolving_slot_restore_test.go — the readers of Game.resolving across
// a restore (#4270). A keyed prompt carries no resume frame, so a table
// waiting on one is a restore point, and the answer's continuation reads
// the resolving slot. The Wandering Archaic copy is
// TestWanderingArchaicCopyConfirmIsARestorePoint; this is a second
// reader. The slot's lifetime across a restore is
// game.TestRestoredResolvingSlotIsClearedByTheNextEventBatch.

// Last Rites' keyed discard, restored while it is open: the cards the
// answer discards are moved BY the resolving spell, so each
// discard names the spell's controller and the spell as its cause, which
// the effect gates (Tamiyo's "can't cause you to discard") and Ranar read.
// Before the slot was carried the restored answer discarded with no cause.
func TestKeyedDiscardAnswerAfterARestoreStillNamesTheResolvingSpellAsItsCause(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { caster.Hand.Cards = nil })
	a := handCardForTest(caster, "Card A", "Sorcery", "")
	b := handCardForTest(caster, "Card B", "Sorcery", "")
	revealHand(victim, rhForest(), rhBolt(), pbBear(), pbDivination())

	castAt(t, g, "Last Rites", "Sorcery", lastRitesOracle, pbPlayer(victim.ID))
	if ask := chooseCardsChoiceFor(g, caster.ID); ask == nil || ask.PickThen != lastRitesReveal.Key() {
		t.Fatalf("the discard is %+v, want the keyed prompt", ask)
	}
	var spell uuid.UUID
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventCast && ev.Actor == caster.ID {
				spell = ev.CardID
			}
		}
	})
	if spell == uuid.Nil {
		t.Fatal("no cast event for Last Rites")
	}

	g = restoreRoundTrip(t, g, true)
	caster = g.Seats[0]
	answerChooseCards(t, g, caster.ID, a, b)

	found := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind != game.EventDiscardCard || (ev.CardID != a && ev.CardID != b) {
				continue
			}
			found++
			if ev.Cause != game.MoveCauseEffect || ev.CauseController != caster.ID || ev.CauseItem != spell {
				t.Errorf("discard of %v: cause %q by %v via %v, want an effect cause by %v via %v",
					ev.CardID, ev.Cause, ev.CauseController, ev.CauseItem, caster.ID, spell)
			}
		}
	})
	if found != 2 {
		t.Errorf("found %d discard events for the two cards, want 2", found)
	}
}
