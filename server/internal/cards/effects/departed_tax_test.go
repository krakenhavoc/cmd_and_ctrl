package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// departed_tax_test.go — #961 / CR 800.4f at the card level: "if an
// object requires a player who has left the game to pay a cost or
// choose whether to pay a cost, that cost is not paid". The three
// pay-unless cards in the catalog, each asked of a player who then
// concedes with the prompt still open.
//
// The two taxes a SURVIVOR's object charges still pay out — the
// payment is what did not happen, not the consequence. The one a
// player's own permanent charges them does nothing at all: CR 800.4a
// takes that permanent out of the game in the same breath.
//
// The engine-side seam (the departure table's drop action, the sweep
// that runs it, undo) is pinned in internal/game's
// leave_game_pay_unless_test.go.

// sacrificedInLog reports whether the log records `id` being
// sacrificed.
func sacrificedInLog(g *game.Game, id uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == game.EventSacrifice && ev.CardID == id {
			return true
		}
	}
	return false
}

// Rhystic Study: the caster concedes owing the {1}. The cost is not
// paid, which opens the Study's controller's own "draw a card?"
// (#1565, MayChoice) exactly as a plain "no" would; answering it
// draws the same card they would have drawn before that choice
// existed.
func TestRhysticStudyDrawsWhenTheTaxedPlayerConcedes(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Rhystic Study", rhysticStudyOracle, "Enchantment")
	handBefore := owner.Hand.Size()

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, caster.ID) {
		t.Fatalf("no pay_unless prompt for the caster after the trigger resolved")
	}

	if err := g.Concede(caster.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if hasPayUnlessFor(g, caster.ID) {
		t.Errorf("the tax is still queued for a player who has left the game")
	}
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != owner.ID {
		t.Fatalf("no 'draw a card?' prompt addressed to the Study's controller")
	}
	if err := g.ResolveConfirm(ask.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveConfirm(draw): %v", err)
	}
	if got := owner.Hand.Size() - handBefore; got != 1 {
		t.Errorf("cards drawn off the conceding caster's unpaid tax = %d, want 1", got)
	}
}

// Smothering Tithe: the drawer concedes owing the {2}, and the Tithe's
// controller still gets the Treasure.
func TestSmotheringTitheMakesItsTreasureWhenTheDrawerConcedes(t *testing.T) {
	g := newCatalogGame(t)
	drawer, owner := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Smothering Tithe", smotheringTitheOracle, "Enchantment")

	// Game.DrawCard is a deliberate no-op during the active seat's own
	// draw step, where newCatalogGame parks the cursor (#692).
	advanceToMain(t, g)
	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, drawer.ID) {
		t.Fatalf("no pay_unless prompt for the drawer")
	}

	if err := g.Concede(drawer.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := countOnBattlefield(g, "Treasure", owner.ID); n != 1 {
		t.Errorf("Treasures under the Tithe's controller after the drawer conceded = %d, want 1", n)
	}
}

// Cumulative upkeep is the other side of the same rule. "Sacrifice
// this permanent unless you pay its upkeep cost" is asked of the
// permanent's OWN controller, so when they concede the permanent
// leaves the game with them (CR 800.4a) instead of being sacrificed —
// a sacrifice is a death, and the rest of the table's triggers watch
// for one.
func TestCumulativeUpkeepSacrificesNothingWhenItsControllerConcedes(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	remora := pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")

	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, owner.ID) {
		t.Fatalf("no cumulative-upkeep prompt for the Remora's controller")
	}

	if err := g.Concede(owner.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if sacrificedInLog(g, remora) {
		t.Errorf("the permanent was sacrificed for an upkeep its controller was no longer there to owe")
	}
	if _, ok := battlefieldCard(g, remora); ok {
		t.Errorf("the departed player's permanent is still on the battlefield")
	}
	if owner.Graveyard != nil && owner.Graveyard.Contains(remora) {
		t.Errorf("the permanent reached a graveyard; it should have left the game (CR 800.4a)")
	}
}
