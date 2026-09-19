package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const roadsideReliquaryOracle = "2fb13687-0518-4ba0-a5ae-dd609464b026"

// crackRoadsideReliquary seeds the land, pays the {2} and activates
// the sacrifice ability, returning the hand size before it resolved.
func crackRoadsideReliquary(t *testing.T, g *game.Game, me *game.Player) int {
	t.Helper()
	land := pushPermanentForTest(g, me.ID, "Roadside Reliquary", roadsideReliquaryOracle, "Land")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Roadside Reliquary: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Errorf("the land should be sacrificed at announce, not at resolution")
	}
	if me.Hand.Size() != before {
		t.Errorf("no card should be drawn before the ability resolves")
	}
	passPriorityAroundTable(t, g)
	return before
}

// TestRoadsideReliquaryDrawsForNeither: the two draws are both
// conditional, and an empty board earns nothing.
func TestRoadsideReliquaryDrawsForNeither(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := crackRoadsideReliquary(t, g, me)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("hand delta %d, want 0 with neither an artifact nor an enchantment", got)
	}
}

// TestRoadsideReliquaryDrawsOneForAnArtifact: the artifact sentence
// alone is one card, not two.
func TestRoadsideReliquaryDrawsOneForAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "A Rock", "", "Artifact")
	before := crackRoadsideReliquary(t, g, me)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("hand delta %d, want 1 for an artifact alone", got)
	}
}

// TestRoadsideReliquaryDrawsOneForAnEnchantment is the mirror: the
// two sentences are independent, so the enchantment half stands on
// its own.
func TestRoadsideReliquaryDrawsOneForAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "A Glow", "", "Enchantment")
	before := crackRoadsideReliquary(t, g, me)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("hand delta %d, want 1 for an enchantment alone", got)
	}
}

// TestRoadsideReliquaryDrawsTwoForBoth is the payoff, and the
// assertion that the two draws really are independent rather than
// one draw gated on both conditions.
func TestRoadsideReliquaryDrawsTwoForBoth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushPermanentForTest(g, me.ID, "A Rock", "", "Artifact")
	pushPermanentForTest(g, me.ID, "A Glow", "", "Enchantment")
	before := crackRoadsideReliquary(t, g, me)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("hand delta %d, want 2 for an artifact and an enchantment", got)
	}
}

// TestRoadsideReliquaryIgnoresOpponentPermanents: "you control" is
// real — an opponent's artifact earns nothing.
func TestRoadsideReliquaryIgnoresOpponentPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, opp.ID, "Their Rock", "", "Artifact")
	pushPermanentForTest(g, opp.ID, "Their Glow", "", "Enchantment")
	before := crackRoadsideReliquary(t, g, me)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("hand delta %d, want 0 — the opponent's board is not yours", got)
	}
}

// TestRoadsideReliquaryTapsForColorless pins the mana half.
func TestRoadsideReliquaryTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushPermanentForTest(g, me.ID, "Roadside Reliquary", roadsideReliquaryOracle, "Land")
	abilities := game.ManaAbilitiesForCard(game.Card{OracleID: roadsideReliquaryOracle})
	if len(abilities) != 1 || abilities[0].Produced != "{C}" {
		t.Fatalf("mana abilities = %+v, want one {C}", abilities)
	}
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}: %v", err)
	}
	if len(me.ManaPool) != 1 || me.ManaPool[0].Color != "C" {
		t.Errorf("pool = %+v, want one {C}", me.ManaPool)
	}
}
