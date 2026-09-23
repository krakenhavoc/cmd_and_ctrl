package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// valakut_exploration_test.go — #1218: the card proof that
// b27ExiledWith plus the existing exile-with-permission and
// exile-to-graveyard primitives are enough to ship this card, with no
// new engine seam. See the doc comment on valakut_exploration.go.

func TestValakutExplorationExilesOnLandfallAndGrantsCastPermission(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedPermanentWithOracle(g, me.ID, "Valakut Exploration", "Enchantment", valakutExplorationOracle)
	loot := stackLibrary(me, "Stashed Bolt", "Instant", "{R}")

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(loot) {
		t.Fatalf("landfall should have exiled the top card of the library")
	}
	perm := exiledPermission(g, loot)
	if !permissionLive(g, perm, me.ID) {
		t.Fatalf("the controller should be able to play the exiled card")
	}
	if err := g.CastSpell(me.ID, loot, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Errorf("casting the exiled card: %v", err)
	}
}

func TestValakutExplorationSweepsUnplayedCardsAtEndStepAndDamagesOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Valakut Exploration", "Enchantment", valakutExplorationOracle)
	loot := stackLibrary(me, "Stashed Bolt", "Instant", "{R}")

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(loot) {
		t.Fatalf("landfall should have exiled the top card of the library")
	}

	oppLifeBefore := opp.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(loot) {
		t.Errorf("the unplayed exiled card should be in its owner's graveyard")
	}
	if got := oppLifeBefore - opp.Life; got != 1 {
		t.Errorf("opponent took %d damage, want 1 (one card swept)", got)
	}
}

func TestValakutExplorationDoesNotSweepACardAlreadyPlayed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Valakut Exploration", "Enchantment", valakutExplorationOracle)
	loot := stackLibrary(me, "Stashed Bolt", "Instant", "{R}")

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, loot, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("casting the exiled card: %v", err)
	}
	passPriorityAroundTable(t, g)

	oppLifeBefore := opp.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	if got := oppLifeBefore - opp.Life; got != 0 {
		t.Errorf("opponent took %d damage for a card that was already played, want 0", got)
	}
}

func TestValakutExplorationDealsNoDamageWithNothingExiled(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedPermanentWithOracle(g, me.ID, "Valakut Exploration", "Enchantment", valakutExplorationOracle)

	oppLifeBefore := opp.Life
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	if got := oppLifeBefore - opp.Life; got != 0 {
		t.Errorf("opponent took %d damage with nothing exiled, want 0", got)
	}
}

// TestValakutExplorationPermissionOutlivesTheEnchantment proves the
// duration choice: "for as long as it remains exiled" is not tied to
// Valakut Exploration's own survival, unlike the sweep that depends on
// it still being there to trigger. Destroying the enchantment before
// its end step removes the thing that would have swept the card, and
// the CAST permission — granted independently, per CastPermission's
// WhileInZone duration — survives regardless.
func TestValakutExplorationPermissionOutlivesTheEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	valakut := seedPermanentWithOracle(g, me.ID, "Valakut Exploration", "Enchantment", valakutExplorationOracle)
	loot := stackLibrary(me, "Stashed Bolt", "Instant", "{R}")

	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(loot) {
		t.Fatalf("landfall should have exiled the top card of the library")
	}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(valakut); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(loot) {
		t.Errorf("nothing should have swept the card once Valakut Exploration was gone")
	}
	perm := exiledPermission(g, loot)
	if !permissionLive(g, perm, me.ID) {
		t.Errorf("the cast permission should survive the enchantment's own departure")
	}
}
