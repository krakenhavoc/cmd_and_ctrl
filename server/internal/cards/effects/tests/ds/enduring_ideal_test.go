package testsds

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const enduringIdealOracle = "d2045055-197c-436a-9874-199db5fe1c96"

// enchantmentsOnBattlefield counts me's enchantments on the battlefield.
func enduringIdealEnchantments(g *game.Game, me *game.Player) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsEnchantment() {
			n++
		}
	}
	return n
}

// enduringIdealUpkeepCopy takes the table to me's next upkeep, resolves the
// epic trigger and its copy, and answers the copy's search with `pick`.
func enduringIdealUpkeepCopy(t *testing.T, g *game.Game, seat int, pickName string) {
	t.Helper()
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	me := g.Seats[seat]
	for i := 0; i < 16 && effectstest.SearchChoiceFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the upkeep copy did not search")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, effectstest.SearchOptionNamed(g, c, pickName))
	effectstest.PassPriorityAroundTable(t, g)
}

// Enduring Ideal (CR 702.50a, 702.50b): the spell finds an enchantment;
// from then on its caster can't cast spells while an opponent still can;
// the ban and the trigger survive a strict restore; and at each of the
// caster's next two upkeeps a copy (not a cast) finds another enchantment,
// with no second epic trigger made by the copies.
func TestEnduringIdealSearchesThenEpicCopiesItEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Ideal A", TypeLine: "Enchantment"},
		game.Card{Name: "Ideal B", TypeLine: "Enchantment"},
		game.Card{Name: "Ideal C", TypeLine: "Enchantment"},
		game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "Enduring Ideal", "Sorcery", enduringIdealOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Enduring Ideal did not search")
	}
	if effectstest.SearchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Fatal("the search offers a creature card")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, effectstest.SearchOptionNamed(g, c, "Ideal A"))
	effectstest.PassPriorityAroundTable(t, g)
	if n := enduringIdealEnchantments(g, me); n != 1 {
		t.Fatalf("%d enchantments after the spell, want 1", n)
	}

	probe := game.Card{Name: "Probe", TypeLine: "Sorcery"}
	if err := effectstest.CastGateFor(g, me.ID, probe, game.ZoneHand); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("my cast after epic: %v, want ErrCantCast", err)
	}
	if err := effectstest.CastGateFor(g, opp.ID, probe, game.ZoneHand); err != nil {
		t.Fatalf("an opponent's cast is refused: %v", err)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[seat]
	if err := effectstest.CastGateFor(g, me.ID, probe, game.ZoneHand); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("my cast after the restore: %v, want ErrCantCast", err)
	}

	enduringIdealUpkeepCopy(t, g, seat, "Ideal B")
	if n := enduringIdealEnchantments(g, me); n != 2 {
		t.Fatalf("%d enchantments after the first upkeep, want 2", n)
	}
	enduringIdealUpkeepCopy(t, g, seat, "Ideal C")
	if n := enduringIdealEnchantments(g, me); n != 3 {
		t.Fatalf("%d enchantments after the second upkeep, want 3", n)
	}
	// CR 707.10: a copy is not cast, so the ban did not stop it and this
	// turn's tally has no spell of mine.
	if got := g.SpellsCastThisTurn[me.ID].Total; got != 0 {
		t.Fatalf("the copy was cast: %d spells of mine this turn", got)
	}
	epics := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.CopiesSpell != nil {
			epics++
		}
	}
	if epics != 1 {
		t.Fatalf("%d epic triggers, want 1: the copies have no epic", epics)
	}
}
