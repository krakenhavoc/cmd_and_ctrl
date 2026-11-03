package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const aprilONeilHacktivistOracle = "d379a919-02cd-46fd-a6ce-b66a77745ae7"

// aprilEndStepDraws puts April on seat 0's battlefield, casts the spells
// with the type lines given (each resolved before the next), walks to
// the end step and resolves April's trigger, and returns how many cards
// it drew.
func aprilEndStepDraws(t *testing.T, typeLines ...string) int {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two", "Three", "Four", "Five", "Six")
	pushCatalogPermanent(g, me.ID, "April O'Neil, Hacktivist", "Legendary Creature — Human Scientist", aprilONeilHacktivistOracle, false)
	toMainForCost(t, g)
	for _, tl := range typeLines {
		id := mayhemHandCard(g, me, "Spell", tl, "{0}", "")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", tl, err)
		}
		passPriorityAroundTable(t, g)
	}
	before := me.Hand.Size()
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	return me.Hand.Size() - before
}

// An artifact creature spell is two card types and a kindred instant two
// more (CR 205.2a, 205.2b, 308.1).
func TestAprilONeilDrawsForEachCardTypeAmongSpellsCast(t *testing.T) {
	if got := aprilEndStepDraws(t, "Artifact Creature — Golem", "Kindred Instant — Elf"); got != 4 {
		t.Errorf("drew %d, want 4 (artifact, creature, instant, kindred)", got)
	}
}

// Two creature spells are one card type.
func TestAprilONeilCountsATypeOnce(t *testing.T) {
	if got := aprilEndStepDraws(t, "Creature — Bear", "Legendary Creature — Elf"); got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestAprilONeilDrawsNothingWithNoSpellCast(t *testing.T) {
	if got := aprilEndStepDraws(t); got != 0 {
		t.Errorf("drew %d with no spell cast, want 0", got)
	}
}
