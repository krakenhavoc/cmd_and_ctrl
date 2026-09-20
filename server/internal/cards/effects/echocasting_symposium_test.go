package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const echocastingSymposiumOracle = "d88c3554-18db-4e5e-a979-2402990a0311"

// The token is the TARGETED player's, not the caster's — the
// political half of the card and the part a two-clause statement
// exists to get right.
func TestEchocastingSymposiumGivesTheTokenToTheTargetedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Push(g, me.ID, "Big Wurm", "Creature — Wurm", "", 6, 6)

	castCatalogSpell(t, g, "Echocasting Symposium", "Sorcery — Lesson", echocastingSymposiumOracle,
		[]game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0},
			{Kind: game.TargetCard, ID: mine, Slot: 1},
		})
	passPriorityAroundTable(t, g)

	var tokens int
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Big Wurm" && IsToken(c) {
			tokens++
			if c.Controller != opp.ID {
				t.Errorf("the copy is the targeted player's, not the caster's (%v)", c.Controller)
			}
			if c.Power != 6 || c.Toughness != 6 {
				t.Errorf("the copy is %d/%d, want 6/6", c.Power, c.Toughness)
			}
		}
	}
	if tokens != 1 {
		t.Errorf("created %d copies, want 1", tokens)
	}
}

// The creature clause is "you control": an opponent's creature is not
// a legal pick, refused at announce (CR 601.2c).
func TestEchocastingSymposiumOnlyCopiesACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)

	err := castCatalogSpellErr(t, g, "Echocasting Symposium", "Sorcery — Lesson", echocastingSymposiumOracle,
		[]game.TargetRef{
			{Kind: game.TargetPlayer, ID: me.ID, Slot: 0},
			{Kind: game.TargetCard, ID: theirs, Slot: 1},
		})
	if err == nil {
		t.Fatal("copying a creature you do not control is refused at announce")
	}
}

// Paradigm is declared as a caveat and nothing pretends to implement
// it: no alternative cost, no self-exile, no delayed trigger.
func TestEchocastingSymposiumDeclaresItsParadigmGap(t *testing.T) {
	spec, ok := Lookup(echocastingSymposiumOracle)
	if !ok || spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Fatalf("one caveat, on a caveats card: %v / %v", spec.Completeness, spec.Caveats)
	}
	if len(spec.AlternativeCosts) != 0 || len(spec.CastableZones) != 0 {
		t.Error("paradigm is deferred, not half-declared as a cost or a cast zone")
	}
}
