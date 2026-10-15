package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gerrards_hourglass_pendant_test.go — Gerrard's Hourglass Pendant (ADR
// 0188 Decision 2, #3560): flash, the shared extra-turn skip, and the
// mass return of the cards put into your graveyard from the battlefield
// this turn.

func seedPendant(g *game.Game, controller uuid.UUID) uuid.UUID {
	return b12Push(g, controller, "Gerrard's Hourglass Pendant", "Legendary Artifact", hourglassPendantOracl, 0, 0)
}

func TestPendantHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	pendant := seedPendant(g, g.Seats[0].ID)
	if !effectiveAbilitiesContain(t, g, pendant, "flash") {
		t.Error("the Pendant does not have flash")
	}
}

func TestPendantSkipsEveryExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	seedPendant(g, g.Seats[me].ID)
	queueOwnExtraTurn(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the Pendant skipped the controller's extra turn")
	if got := skippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
}

func destroyNow(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

// The set is exactly the artifact, creature, enchantment and land cards
// put into YOUR graveyard from the battlefield THIS turn. A planeswalker
// is not on the list, a card that arrived from the hand is not, a card
// that died last turn is not, and an opponent's graveyard is untouched.
// The Pendant is exiled as a cost, so it is not among them.
func TestPendantReturnsOnlyWhatWasPutThereFromTheBattlefieldThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pendant := seedPendant(g, me.ID)

	// Last turn's casualty: it stays.
	old := b12Creature(g, me.ID, "Old Bear", "Creature — Bear", 2, 2)
	destroyNow(t, g, old)
	endTurn(t, g)
	endTurn(t, g)
	endTurn(t, g)
	endTurn(t, g)
	if g.Turn.ActiveSeat != me.Seat {
		t.Fatalf("not back to my turn: seat %d", g.Turn.ActiveSeat)
	}
	advanceToMain(t, g)

	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	ench := b12Push(g, me.ID, "Aura", "Enchantment", "", 0, 0)
	land := b12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	walker := b12Push(g, me.ID, "Walker", "Legendary Planeswalker — Test", "", 0, 0)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{bear, rock, ench, land, walker, theirs} {
		destroyNow(t, g, id)
	}
	fromHand := pushGraveyardCardTyped(me, "Discarded Bear", "Creature — Bear")
	for _, id := range []uuid.UUID{bear, rock, ench, land, walker} {
		if !me.Graveyard.Contains(id) {
			t.Fatalf("setup: a card did not reach the graveyard")
		}
	}

	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, pendant, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(pendant) || g.Battlefield.Contains(pendant) {
		t.Fatal("the Pendant is not in exile once its cost is paid")
	}
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"creature": bear, "artifact": rock, "enchantment": ench, "land": land} {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Errorf("the %s did not return", name)
			continue
		}
		if !c.Tapped {
			t.Errorf("the %s returned untapped", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"planeswalker": walker, "card from hand": fromHand, "last turn's creature": old} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("the %s left the graveyard", name)
		}
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's graveyard was touched")
	}
}

// Nothing qualifies: the ability still resolves, the Pendant is still gone.
func TestPendantWithNothingToReturnStillExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pendant := seedPendant(g, me.ID)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, pendant, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(pendant) {
		t.Error("the Pendant was not exiled")
	}
}
