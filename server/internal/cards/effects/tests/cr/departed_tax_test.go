package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhystic Study: the caster concedes owing the {1}. The cost is not
// paid, which opens the Study's controller's own "draw a card?"
// (#1565, MayChoice) exactly as a plain "no" would; answering it
// draws the same card they would have drawn before that choice
// existed.
func TestRhysticStudyDrawsWhenTheTaxedPlayerConcedes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
	handBefore := owner.Hand.Size()

	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, caster.ID) {
		t.Fatalf("no pay_unless prompt for the caster after the trigger resolved")
	}

	if err := g.Concede(caster.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if effectstest.HasPayUnlessFor(g, caster.ID) {
		t.Errorf("the tax is still queued for a player who has left the game")
	}
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
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
	g := effectstest.NewCatalogGame(t)
	drawer, owner := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Smothering Tithe", effectstest.SmotheringTitheOracle, "Enchantment")

	// Game.DrawCard is a deliberate no-op during the active seat's own
	// draw step, where newCatalogGame parks the cursor (#692).
	effectstest.AdvanceToMain(t, g)
	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatalf("DrawCard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, drawer.ID) {
		t.Fatalf("no pay_unless prompt for the drawer")
	}

	if err := g.Concede(drawer.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", owner.ID); n != 1 {
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
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	remora := effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, owner.ID) {
		t.Fatalf("no cumulative-upkeep prompt for the Remora's controller")
	}

	if err := g.Concede(owner.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if effectstest.SacrificedInLog(g, remora) {
		t.Errorf("the permanent was sacrificed for an upkeep its controller was no longer there to owe")
	}
	if _, ok := effectstest.BattlefieldCard(g, remora); ok {
		t.Errorf("the departed player's permanent is still on the battlefield")
	}
	if owner.Graveyard != nil && owner.Graveyard.Contains(remora) {
		t.Errorf("the permanent reached a graveyard; it should have left the game (CR 800.4a)")
	}
}
