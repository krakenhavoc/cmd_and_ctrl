package effects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hybrid_mana_value_test.go pins CR 202.3f at the card level: a
// monocoloured hybrid symbol ({2/W}) is worth its larger component,
// two, to every "mana value N or less" reading in the catalog. Both
// of the catalog's shapes are covered — the ManaValueLE target /
// match predicate (Ritual of Soot) and the Card.ManaValue read a
// search predicate closes over (Beseech the Queen).

// TestRitualOfSootReadsMonocolouredHybridAsTwo: a {2/G}{2/G} creature
// is mana value 4 and survives "destroy all creatures with mana value
// 3 or less"; a {W/U}{W/U} creature is mana value 2 and does not.
func TestRitualOfSootReadsMonocolouredHybridAsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	twoColour := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: twoColour, Name: "Two-Colour Hybrid", TypeLine: "Creature — Elemental",
		ManaCost: "{W/U}{W/U}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	monocoloured := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: monocoloured, Name: "Wildgrowth-Style Archaic", TypeLine: "Creature — Avatar",
		ManaCost: "{2/G}{2/G}", Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID,
	})

	castCatalogSpell(t, g, "Ritual of Soot", "Sorcery", ritualOfSootOracle, nil)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(twoColour) {
		t.Error("a {W/U}{W/U} creature (mana value 2) survived Ritual of Soot")
	}
	if !g.Battlefield.Contains(monocoloured) {
		t.Error("a {2/G}{2/G} creature (mana value 4, CR 202.3f) was destroyed by Ritual of Soot — its hybrid symbols were read as 1")
	}
}

// TestBeseechTheQueenReadsMonocolouredHybridAsTwo: with three lands,
// Spectral Procession ({2/W}{2/W}{2/W}, mana value 6) is out of reach
// and a {1}{W/U}{W/U} card (mana value 3) is not. Before CR 202.3f was
// honoured both were offered — a tutor stronger than printed.
func TestBeseechTheQueenReadsMonocolouredHybridAsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		pushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	}
	three := pushLibraryCardForTest(me, game.Card{
		Name: "Three With Hybrid", TypeLine: "Creature — Bird", ManaCost: "{1}{W/U}{W/U}",
	})
	procession := pushLibraryCardForTest(me, game.Card{
		Name: "Spectral Procession", TypeLine: "Sorcery", ManaCost: "{2/W}{2/W}{2/W}",
	})

	castCatalogSpell(t, g, "Beseech the Queen", "Sorcery", beseechOracle, nil)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.SearchCards {
		offered[id] = true
	}
	if !offered[three] {
		t.Error("a {1}{W/U}{W/U} card (mana value 3) was not offered with three lands out")
	}
	if offered[procession] {
		t.Error("Spectral Procession (mana value 6, CR 202.3f) was offered with three lands out")
	}
	answerSearchByID(t, g, me.ID, three)
}
