package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// world_war_hulk_test.go — World War Hulk (Marvel Super Heroes, #4455, #4615):
// the colour clause on a next-spell promise (PermissionFilter.ColorsAny) and
// each chapter, played through the real catalog entry.

const worldWarHulkOracle = "8c4da760-8913-46ca-b00c-bba6d877d482"

// worldWarHulkTable casts the Saga for seat 0 and settles chapter I.
func worldWarHulkTable(t *testing.T) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	saga := castCatalogSpell(t, g, "World War Hulk", "Enchantment — Saga", worldWarHulkOracle, nil)
	passPriorityAroundTable(t, g)
	return g, saga
}

// worldWarHulkHandCard seeds a coloured card into seat 0's hand.
func worldWarHulkHandCard(me *game.Player, name, typeLine, manaCost string, colors ...string) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, Colors: colors,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	return id
}

func worldWarHulkFree(g *game.Game, me *game.Player, card uuid.UUID) bool {
	return offerWithKey(grantedCardOffers(g, me.ID, card, game.ZoneHand), game.GrantedAltCostNextFree) != nil
}

func TestWorldWarHulkIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(worldWarHulkOracle)
	if !ok {
		t.Fatal("World War Hulk is not registered")
	}
	if spec.Name != "World War Hulk" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("spec = %q / %v / %v, want World War Hulk, Full, no caveats", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// Chapter I: red and green creature spells are offered the free cast; a blue
// creature, a colourless creature and a red noncreature are not.
func TestWorldWarHulkChapterIOffersRedAndGreenCreatureSpellsTheFreeCast(t *testing.T) {
	g, _ := worldWarHulkTable(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, cost string
		colors               []string
		want                 bool
	}{
		{"Red Beast", "Creature — Beast", "{4}{R}", []string{"R"}, true},
		{"Green Beast", "Creature — Beast", "{4}{G}", []string{"G"}, true},
		{"Gruul Beast", "Creature — Beast", "{3}{R}{G}", []string{"R", "G"}, true},
		{"Simic Beast", "Creature — Beast", "{3}{G}{U}", []string{"G", "U"}, true},
		{"Blue Beast", "Creature — Beast", "{4}{U}", []string{"U"}, false},
		{"Izzet Beast", "Creature — Beast", "{3}{U}{B}", []string{"U", "B"}, false},
		{"Golem", "Artifact Creature — Golem", "{5}", nil, false},
		{"Red Sorcery", "Sorcery", "{4}{R}", []string{"R"}, false},
		{"Green Instant", "Instant", "{G}", []string{"G"}, false},
		{"Red Artifact Creature", "Artifact Creature — Golem", "{4}{R}", []string{"R"}, true},
	} {
		id := worldWarHulkHandCard(me, tc.name, tc.typeLine, tc.cost, tc.colors...)
		if got := worldWarHulkFree(g, me, id); got != tc.want {
			t.Errorf("%s: free cast offered = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// The free cast is real: with an empty pool under the strict mana gate a red
// creature spell is cast, a blue one is refused for want of mana.
func TestWorldWarHulkFreeCastPaysNothingAndOtherSpellsStillPay(t *testing.T) {
	g, _ := worldWarHulkTable(t)
	me := g.Seats[0]
	blue := worldWarHulkHandCard(me, "Blue Beast", "Creature — Beast", "{4}{U}", "U")
	if err := g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a blue creature spell was cast with no mana")
	}
	red := worldWarHulkHandCard(me, "Red Beast", "Creature — Beast", "{4}{R}", "R")
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostNextFree}); err != nil {
		t.Fatalf("free red creature cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := ceOnBattlefield(t, g, red); c.Name != "Red Beast" {
		t.Errorf("battlefield card = %q, want Red Beast", c.Name)
	}
}

func TestWorldWarHulkFreeGreenCreatureCast(t *testing.T) {
	g, _ := worldWarHulkTable(t)
	me := g.Seats[0]
	green := worldWarHulkHandCard(me, "Green Beast", "Creature — Beast", "{5}{G}", "G")
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostNextFree}); err != nil {
		t.Fatalf("free green creature cast: %v", err)
	}
}

// One promise, one spend: the first matching spell takes it, a spell the
// filter does not match leaves it, and a second matching spell pays.
func TestWorldWarHulkPromiseIsSpentOnceByTheFirstMatchingSpell(t *testing.T) {
	g, _ := worldWarHulkTable(t)
	me := g.Seats[0]

	blue := worldWarHulkHandCard(me, "Blue Beast", "Creature — Beast", "{4}{U}", "U")
	if err := g.CastSpell(me.ID, blue, game.CastSpellParams{}); err != nil {
		t.Fatalf("paying for the blue creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	sorcery := worldWarHulkHandCard(me, "Red Rite", "Sorcery", "{4}{R}", "R")
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
		t.Fatalf("paying for the red sorcery: %v", err)
	}
	passPriorityAroundTable(t, g)
	red := worldWarHulkHandCard(me, "Red Beast", "Creature — Beast", "{4}{R}", "R")
	if !worldWarHulkFree(g, me, red) {
		t.Fatal("a blue creature and a red sorcery used up the promise")
	}
	if err := g.CastSpell(me.ID, red, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostNextFree}); err != nil {
		t.Fatalf("free red creature cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	green := worldWarHulkHandCard(me, "Green Beast", "Creature — Beast", "{4}{G}", "G")
	if worldWarHulkFree(g, me, green) {
		t.Error("the promise outlived the first red or green creature spell")
	}
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{Strict: true, AlternativeCost: game.GrantedAltCostNextFree}); err == nil {
		t.Error("a second creature spell was cast for free")
	}
}

// "This turn": a promise nobody used lapses at end of turn.
func TestWorldWarHulkPromiseLapsesAtEndOfTurn(t *testing.T) {
	g, _ := worldWarHulkTable(t)
	me := g.Seats[0]
	early := worldWarHulkHandCard(me, "Early Beast", "Creature — Beast", "{4}{R}", "R")
	if !worldWarHulkFree(g, me, early) {
		t.Fatal("the red creature was not offered the free cast this turn")
	}
	nextTurnOfMine(t, g)
	late := worldWarHulkHandCard(me, "Late Beast", "Creature — Beast", "{4}{R}", "R")
	if worldWarHulkFree(g, me, late) {
		t.Error("the free cast outlived the turn it was granted")
	}
}

// worldWarHulkNextChapter walks to the next precombat main of seat 0, where
// the Saga's lore counter arrives, and answers the chapter's target.
func worldWarHulkNextChapter(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	advanceToPrecombatMainOf(t, g, 0)
	if latestPickTarget(g, g.Seats[0].ID) != nil {
		pickCard(t, g, g.Seats[0].ID, target)
	}
	passPriorityAroundTable(t, g)
}

func TestWorldWarHulkChapterIIPutsThreeCountersOnTheChosenCreature(t *testing.T) {
	g, _ := worldWarHulkTable(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	other := pushVanillaCreature(g, me.ID, "My Other Bear", 3, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	worldWarHulkNextChapter(t, g, mine)
	if got := ceOnBattlefield(t, g, mine).Counters[game.CounterPlusOne]; got != 3 {
		t.Errorf("chosen creature has %d +1/+1 counters, want 3", got)
	}
	if got := ceOnBattlefield(t, g, other).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("another creature of mine got %d counters", got)
	}
	if got := ceOnBattlefield(t, g, theirs).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestWorldWarHulkChapterIIOnlyOffersMyCreatures(t *testing.T) {
	g, _ := worldWarHulkTable(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	pushVanillaCreature(g, me.ID, "My Other Bear", 3, 3)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToPrecombatMainOf(t, g, 0)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter II asked for no target")
	}
	if !hasID(p.PickTargetCards, mine) || hasID(p.PickTargetCards, theirs) {
		t.Errorf("legal targets %v: want my creatures, never %s", p.PickTargetCards, theirs)
	}
}

func TestWorldWarHulkChapterIIIDoublesPowerAndToughnessAndGivesTrampleThenSacrifices(t *testing.T) {
	g, saga := worldWarHulkTable(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "My Bear", 3, 4)
	worldWarHulkNextChapter(t, g, bear) // II: three counters, so 6/7
	if c := ceOnBattlefield(t, g, bear); c.CurrentPower() != 6 || c.CurrentToughness() != 7 {
		t.Fatalf("after chapter II the bear is %d/%d, want 6/7", c.CurrentPower(), c.CurrentToughness())
	}
	if c := ceOnBattlefield(t, g, bear); game.HasKeyword(&c, "trample") {
		t.Fatal("the bear had trample before chapter III")
	}
	worldWarHulkNextChapter(t, g, bear) // III: doubled, 12/14 and trample
	c := ceOnBattlefield(t, g, bear)
	if c.CurrentPower() != 12 || c.CurrentToughness() != 14 {
		t.Errorf("after chapter III the bear is %d/%d, want 12/14", c.CurrentPower(), c.CurrentToughness())
	}
	if !game.HasKeyword(&c, "trample") {
		t.Error("the bear did not gain trample")
	}
	if !inGraveyardOf(g, me.ID, saga) {
		t.Error("the Saga should be sacrificed after chapter III")
	}
	advanceToNextSeatsTurn(t, g)
	c = ceOnBattlefield(t, g, bear)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 7 || game.HasKeyword(&c, "trample") {
		t.Errorf("after end of turn the bear is %d/%d (trample %v), want 6/7 and no trample",
			c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "trample"))
	}
}
