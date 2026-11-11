package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const voidAttendantOracle = "34f3ec68-871f-4b94-b497-04f71aea88c9"

func voidAttendantScions(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && IsToken(c) && c.Name == "Eldrazi Scion" {
			n++
		}
	}
	return n
}

// The cost puts an opponent's exiled card into ITS OWNER's graveyard, and
// the ability makes one Scion. Nothing else pays it: not the activator's
// own exiled card, not a card outside exile, and not nothing.
func TestVoidAttendantProcessesAnOpponentsExiledCardForAScion(t *testing.T) {
	g, me, opp := exileCostTable(t)
	attendant := pushCatalogPermanent(g, me.ID, "Void Attendant", "Creature — Eldrazi Processor", voidAttendantOracle, false)
	mine := pushExiledCard(g, me.ID, "My Exiled Bear", "Creature — Bear", false)
	theirs := pushExiledCard(g, opp.ID, "Their Exiled Bolt", "Instant", false)
	var inYard uuid.UUID
	g.WithWriteLock(func() { inYard = pushGraveyardCardTyped(opp, "Their Dead Bear", "Creature — Bear") })
	floatMana(t, g, me, "{G}{G}{G}{G}")

	for name, ids := range map[string][]uuid.UUID{
		"the activator's own exiled card": {mine},
		"a card in a graveyard":           {inYard},
		"no card":                         nil,
		"the same card twice":             {theirs, theirs},
	} {
		if err := g.ActivateCatalogAbility(me.ID, attendant, 0, game.ActivateAbilityParams{ExileIDs: ids}); err == nil {
			t.Fatalf("%s paid \"put a card an opponent owns from exile into that player's graveyard\"", name)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, attendant, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{theirs}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Exile.Contains(theirs) || !opp.Graveyard.Contains(theirs) {
		t.Fatal("the opponent's card did not go from exile to its owner's graveyard")
	}
	if !g.Exile.Contains(mine) {
		t.Fatal("the activator's own exiled card moved")
	}
	noDiscardEvents(t, g)
	passPriorityAroundTable(t, g)
	if n := voidAttendantScions(g, me.ID); n != 1 {
		t.Fatalf("%d Eldrazi Scion tokens, want 1", n)
	}

	// With no opponent's card left in exile the cost can't be paid.
	if err := g.ActivateCatalogAbility(me.ID, attendant, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{mine}}); err == nil {
		t.Fatal("the ability was activated with no opponent's card in exile")
	}
	var options []uuid.UUID
	g.WithWriteLock(func() {
		options = g.ExileCostOptionsForEffect(me.ID, attendant, ProcessCards(1, "a card", nil).ExileCards)
	})
	if len(options) != 0 {
		t.Fatalf("the walk still offers %v", options)
	}
}

// A face-down card is a card an opponent owns, so it pays the cost; the
// player has only picked its group, and one card of the group moves.
func TestVoidAttendantProcessesOneFaceDownCardOfTheGroup(t *testing.T) {
	g, me, opp := exileCostTable(t)
	attendant := pushCatalogPermanent(g, me.ID, "Void Attendant", "Creature — Eldrazi Processor", voidAttendantOracle, false)
	a := pushExiledCard(g, opp.ID, "Hidden A", "Sorcery", true)
	b := pushExiledCard(g, opp.ID, "Hidden B", "Sorcery", true)
	faceUp := pushExiledCard(g, opp.ID, "Seen", "Sorcery", false)
	floatMana(t, g, me, "{G}{G}")

	if err := g.ActivateCatalogAbility(me.ID, attendant, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{a}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	moved := 0
	for _, id := range []uuid.UUID{a, b} {
		if opp.Graveyard.Contains(id) {
			moved++
			if c, ok := g.LookupCardForEffect(id); !ok || c.FaceDown {
				t.Error("the processed card is still face down in the graveyard")
			}
		}
	}
	if moved != 1 || !g.Exile.Contains(faceUp) {
		t.Fatalf("%d face-down cards moved (face-up card still exiled: %v), want exactly one of the group", moved, g.Exile.Contains(faceUp))
	}
}
