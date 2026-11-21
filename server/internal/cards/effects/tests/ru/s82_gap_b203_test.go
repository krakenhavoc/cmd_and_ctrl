package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b203_test.go — cards missing from sets since AFR, batch 203
// (#5948). Each test plays the real catalog entry and asserts what its text
// says.

const (
	gapB203Murkfiend = "61d28182-498f-4bbc-bb7a-c5e1ef872dda"
	gapB203Ruric     = "6ed13a89-4284-4c01-8099-75e2d68c8c63"
	gapB203Chronicle = "c3779271-1186-4754-8b8d-0ee18e585bb6"
)

func TestGapB203EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB203Murkfiend: "Murkfiend Liege",
		gapB203Ruric:     "Ruric Thar, the Unbowed",
		gapB203Chronicle: "Chronicler of Heroes",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// gapB203Colour sets a battlefield card's colours (a pushed test card has no mana cost).
func gapB203Colour(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Colors = colors
			}
		}
		g.BumpLayerVersionForTest()
	})
}

// gapB203Bear puts a 2/2 of the given colours onto the battlefield.
func gapB203Bear(g *game.Game, owner uuid.UUID, colors ...string) uuid.UUID {
	id := effectstest.B12Creature(g, owner, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Colors = colors
			}
		}
		g.BumpLayerVersionForTest()
	})
	return id
}

// --- Murkfiend Liege ---------------------------------------------------------

func TestGapB203MurkfiendLiegeGivesGreenAndBlueCreaturesPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := effectstest.B12Push(g, me.ID, "Murkfiend Liege", "Creature — Horror", gapB203Murkfiend, 4, 4)
	green, blue, both := gapB203Bear(g, me.ID, "G"), gapB203Bear(g, me.ID, "U"), gapB203Bear(g, me.ID, "G", "U")
	red, theirs := gapB203Bear(g, me.ID, "R"), gapB203Bear(g, opp.ID, "G")
	for _, want := range []struct {
		id  uuid.UUID
		who string
		p   int
		t   int
	}{
		{liege, "the Liege", 4, 4}, {green, "green", 3, 3}, {blue, "blue", 3, 3}, {both, "green and blue", 4, 4},
		{red, "red", 2, 2}, {theirs, "an opponent's green creature", 2, 2},
	} {
		if p, tough := effectstest.PtOf(t, g, want.id); p != want.p || tough != want.t {
			t.Errorf("%s is %d/%d, want %d/%d", want.who, p, tough, want.p, want.t)
		}
	}
}

func TestGapB203MurkfiendLiegeUntapsGreenAndBlueCreaturesInOtherPlayersUntapSteps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	liege := effectstest.B12Push(g, me.ID, "Murkfiend Liege", "Creature — Horror", gapB203Murkfiend, 4, 4)
	gapB203Colour(g, liege, "G", "U")
	green, blue := gapB203Bear(g, me.ID, "G"), gapB203Bear(g, me.ID, "U")
	red := gapB203Bear(g, me.ID, "R")
	notMine := gapB203Bear(g, opp.ID, "G")
	land := effectstest.B12Permanent(g, me.ID, "My Forest", "Land")
	effectstest.TapForTest(t, g, liege, green, blue, red, notMine, land)

	// The opponent's untap step: my green and blue creatures untap, nothing else of mine,
	// and the opponent's own permanents follow their own untap step.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		id         uuid.UUID
		name       string
		wantTapped bool
	}{
		{liege, "the Liege", false}, {green, "my green creature", false}, {blue, "my blue creature", false},
		{red, "my red creature", true}, {land, "my land", true}, {notMine, "the opponent's creature", false},
	} {
		if got := effectstest.B16Tapped(t, g, c.id); got != c.wantTapped {
			t.Errorf("%s tapped=%v after the opponent's untap step, want %v", c.name, got, c.wantTapped)
		}
	}
}

// --- Ruric Thar, the Unbowed -------------------------------------------------

func TestGapB203RuricTharHitsEveryPlayerWhoCastsANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ruric := effectstest.B12Push(g, me.ID, "Ruric Thar, the Unbowed", "Legendary Creature — Ogre Warrior", gapB203Ruric, 6, 6)
	for _, kw := range []string{"reach", "vigilance"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, ruric, kw) {
			t.Errorf("Ruric Thar does not have %s", kw)
		}
	}
	if got := len(effectstest.FindBattlefieldCardForTest(g, ruric).Effective().AttackRequirements); got != 1 {
		t.Errorf("Ruric Thar has %d attack requirements, want 1 (attacks each combat)", got)
	}

	// An opponent's noncreature spell: that player takes 6.
	life := opp.Life
	effectstest.B10OpponentCastsBolt(t, g, opp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Errorf("the caster's life %d -> %d, want -6", life, opp.Life)
	}

	// A creature spell does nothing.
	effectstest.AdvanceToMain(t, g)
	life = me.Life
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if me.Life != life {
		t.Errorf("a creature spell cost its caster %d life", life-me.Life)
	}

	// My own noncreature spell hits me.
	mine := me.Life
	effectstest.B12Bolt(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine-6 {
		t.Errorf("my life %d -> %d, want -6 for my own noncreature spell", mine, me.Life)
	}
}

// --- Chronicler of Heroes ----------------------------------------------------

func TestGapB203ChroniclerDrawsOnlyWithACounteredCreature(t *testing.T) {
	for _, tc := range []struct {
		name    string
		counter string
		draws   int
	}{{"a +1/+1 counter", game.CounterPlusOne, 1}, {"no counter", "", 0}, {"a different counter", game.CounterMinusOne, 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
		if tc.counter != "" {
			g.WithWriteLock(func() {
				for i := range g.Battlefield.Cards {
					if g.Battlefield.Cards[i].InstanceID == friend {
						g.Battlefield.Cards[i].Counters = map[string]int{tc.counter: 1}
					}
				}
			})
		}
		hand := me.Hand.Size()
		effectstest.Msc17Play(t, g, me, game.Card{Name: "Chronicler of Heroes", TypeLine: "Creature — Centaur Wizard", OracleID: gapB203Chronicle, Power: 3, Toughness: 3})
		if got := me.Hand.Size() - hand; got != tc.draws {
			t.Errorf("with %s the hand changed by %d, want %d", tc.name, got, tc.draws)
		}
	}
}
