package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const drownyardTempleOracle = "c30f9be4-c274-4ad0-b5d7-7d3421aa4277"

// TestDrownyardTempleTapsForColorless is the mana half — a plain
// colorless tap, no different from a basic.
func TestDrownyardTempleTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Drownyard Temple", "Land", drownyardTempleOracle, false)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want [C]", got)
	}
}

// TestDrownyardTempleReturnsFromGraveyardTapped is #1284's other
// card: "{3}: Return this card from your graveyard to the
// battlefield tapped." Same shared body as Reassembling Skeleton
// (returnThisFromGraveyardTapped), and the same tap-EVENT discipline
// as that test — see its doc comment for why counting EventTapCard
// rather than reading Tapped is the assertion that matters.
func TestDrownyardTempleReturnsFromGraveyardTapped(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Drownyard Temple", "Land",
		drownyardTempleOracle, 0, 0, "{C}{C}{C}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the temple is not on the battlefield")
	}
	if me.Graveyard.Contains(id) {
		t.Error("the temple is still in the graveyard")
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("the returned temple cannot be looked up")
	}
	if !c.Tapped {
		t.Error("the temple must enter tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("the temple entered via a replacement, not an OnETB tap — got %d EventTapCard, want 0", n)
	}
}

// TestDrownyardTempleGraveyardAbilityIsZoneScoped mirrors
// TestReassemblingSkeletonActivatesFromTheGraveyardOnly for the
// second ability's own zone declaration and cost shape — a mana
// activation and a graveyard activation on the same card must not be
// confused with each other by AbilityFunctionsFromZone.
func TestDrownyardTempleGraveyardAbilityIsZoneScoped(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: drownyardTempleOracle})
	if len(abilities) != 1 {
		t.Fatalf("%d activated abilities, want the one graveyard ability (the mana ability is separate)", len(abilities))
	}
	ab := abilities[0]
	if ab.Cost.Mana != "{3}" {
		t.Errorf("cost %q, want {3}", ab.Cost.Mana)
	}
	if ab.Cost.Tap {
		t.Error("the graveyard ability has no tap component — the land is not on the battlefield to tap")
	}
	if !game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
		t.Error("the ability does not function from a graveyard")
	}
	if game.AbilityFunctionsFromZone(ab, game.ZoneBattlefield) {
		t.Error("the graveyard-return ability must not also be offered on the battlefield")
	}
}
