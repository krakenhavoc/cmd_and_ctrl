package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const neoformOracle = "420c6dcf-966d-4a4c-a0ef-23037ab8b325"

func neoformCast(t *testing.T, g *game.Game, fodder uuid.UUID) {
	t.Helper()
	effectstest.PaCast(t, g, "Neoform", "Sorcery", "{G}{U}", neoformOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{fodder},
	})
	effectstest.PassPriorityAroundTable(t, g)
}

// Neoform finds a creature of EXACTLY one more mana value than the
// sacrificed creature, and it enters with a +1/+1 counter.
func TestNeoformFindsExactlyOneMoreAndAddsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	same := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Same", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	bigger := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bigger", TypeLine: "Creature — Beast", ManaCost: "{3}{G}", Power: 4, Toughness: 4})
	fits := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Fits", TypeLine: "Creature — Elk", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	neoformCast(t, g, effectstest.PaPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2))
	// Only one card has mana value 3, so the pick is forced.
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, fits)
	}
	for _, bad := range []uuid.UUID{same, bigger} {
		if effects.OnBattlefield(g, bad) {
			t.Fatal("a creature with the wrong mana value was found")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, fits) {
		t.Fatal("the mana value 3 creature did not enter")
	}
	if got := effectstest.CountersOn(g, fits, "+1/+1"); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// A token sacrificed has mana value 0, so Neoform looks for mana value 1.
func TestNeoformSacrificedTokenSearchesForManaValueOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	one := effectstest.PushLibraryCardForTest(me, game.Card{Name: "One", TypeLine: "Creature — Elf", ManaCost: "{G}", Power: 1, Toughness: 1})
	neoformCast(t, g, effectstest.PaPermanent(g, me.ID, "Token", "Creature — Spirit", "", 1, 1))
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, one)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, one) {
		t.Fatal("the mana value 1 creature did not enter")
	}
}

// The counter goes through the CR 614 counter pipeline: a Doubling
// Season on the table makes it two.
func TestNeoformCounterIsDoubledByDoublingSeason(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedReplacementPermanent(g, effectstest.DoublingSeasonOracle, "Doubling Season", me.ID)
	fits := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Fits", TypeLine: "Creature — Elk", ManaCost: "{2}{G}", Power: 3, Toughness: 3})
	neoformCast(t, g, effectstest.PaPermanent(g, me.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2))
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		effectstest.AnswerSearchByID(t, g, me.ID, fits)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, fits, "+1/+1"); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2 under Doubling Season", got)
	}
}
