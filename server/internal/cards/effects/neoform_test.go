package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const neoformOracle = "420c6dcf-966d-4a4c-a0ef-23037ab8b325"

func neoformCast(t *testing.T, g *game.Game, fodder uuid.UUID) {
	t.Helper()
	paCast(t, g, "Neoform", "Sorcery", "{G}{U}", neoformOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{fodder},
	})
	passPriorityAroundTable(t, g)
}

// Neoform finds a creature of EXACTLY one more mana value than the
// sacrificed creature, and it enters with a +1/+1 counter.
func TestNeoformFindsExactlyOneMoreAndAddsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	same := pushLibraryCardForTest(me, game.Card{Name: "Same", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	bigger := pushLibraryCardForTest(me, game.Card{Name: "Bigger", TypeLine: "Creature — Beast", ManaCost: "{3}{G}", Power: 4, Toughness: 4})
	fits := pushLibraryCardForTest(me, game.Card{Name: "Fits", TypeLine: "Creature — Elk", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	neoformCast(t, g, paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2))
	// Only one card has mana value 3, so the pick is forced.
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchByID(t, g, me.ID, fits)
	}
	for _, bad := range []uuid.UUID{same, bigger} {
		if onBattlefield(g, bad) {
			t.Fatal("a creature with the wrong mana value was found")
		}
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, fits) {
		t.Fatal("the mana value 3 creature did not enter")
	}
	if got := countersOn(g, fits, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// A token sacrificed has mana value 0, so Neoform looks for mana value 1.
func TestNeoformSacrificedTokenSearchesForManaValueOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	one := pushLibraryCardForTest(me, game.Card{Name: "One", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1})
	neoformCast(t, g, paPermanent(g, me.ID, "Token", "Creature — Spirit", "", 1, 1))
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchByID(t, g, me.ID, one)
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, one) {
		t.Fatal("the mana value 1 creature did not enter")
	}
}

// The counter goes through the CR 614 counter pipeline: a Doubling
// Season on the table makes it two.
func TestNeoformCounterIsDoubledByDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", me.ID)
	fits := pushLibraryCardForTest(me, game.Card{Name: "Fits", TypeLine: "Creature — Elk", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	neoformCast(t, g, paPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2))
	if searchChoiceFor(g, me.ID) != nil {
		answerSearchByID(t, g, me.ID, fits)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, fits, "+1/+1"); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 under Doubling Season", got)
	}
}
