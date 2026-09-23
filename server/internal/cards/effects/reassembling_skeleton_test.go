package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const reassemblingSkeletonOracle = "9dbc3530-b278-4c8d-b2cc-a09dfac9d5e5"

// TestReassemblingSkeletonReturnsFromGraveyardTapped is #1284: the
// card's whole printed ability is "{1}{B}: Return this card from your
// graveyard to the battlefield tapped." The assertion that earns its
// keep is the tap EVENT count, not the Tapped flag — the same
// discipline temples_test.go uses for the Temple cycle's self-entry
// replacement: an OnETB tap would leave the permanent Tapped too, so
// only counting EventTapCard tells a replacement (zero taps, entered
// tapped) from the enters-then-gets-tapped workaround AGENTS.md warns
// against.
func TestReassemblingSkeletonReturnsFromGraveyardTapped(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Reassembling Skeleton", "Creature — Skeleton Warrior",
		reassemblingSkeletonOracle, 1, 1, "{B}{B}", game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the skeleton is not on the battlefield")
	}
	if me.Graveyard.Contains(id) {
		t.Error("the skeleton is still in the graveyard")
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("the returned skeleton cannot be looked up")
	}
	if !c.Tapped {
		t.Error("the skeleton must enter tapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("the skeleton entered via a replacement, not an OnETB tap — got %d EventTapCard, want 0", n)
	}
}

// TestReassemblingSkeletonActivatesFromTheGraveyardOnly pins the CR
// 113.6 zone declaration and the boot-time cost shape, the same
// invariant TestGraveyardKeywordsDeclareAGraveyardAbility asserts for
// the four keyword abilities one file over — Reassembling Skeleton
// declares no keyword, so it needs its own copy of the same check.
func TestReassemblingSkeletonActivatesFromTheGraveyardOnly(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: reassemblingSkeletonOracle})
	if len(abilities) != 1 {
		t.Fatalf("%d activated abilities, want the one graveyard ability", len(abilities))
	}
	ab := abilities[0]
	if ab.Cost.Mana != "{1}{B}" {
		t.Errorf("cost %q, want {1}{B}", ab.Cost.Mana)
	}
	if !game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
		t.Error("the ability does not function from a graveyard")
	}
	if game.AbilityFunctionsFromZone(ab, game.ZoneBattlefield) {
		t.Error("the ability is offered on the battlefield too — nothing printed restricts it there, " +
			"but the card is a creature and this ability only makes sense off the battlefield")
	}
	if ab.SorcerySpeed {
		t.Error("no \"activate only as a sorcery\" is printed on this card, unlike Unearth")
	}
}
