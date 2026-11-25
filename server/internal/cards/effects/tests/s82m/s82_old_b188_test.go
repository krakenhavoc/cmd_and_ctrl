package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b188_test.go — Guilds of Ravnica, batch 188 (#5525): undergrowth.
// Each test plays the real catalog entry and asserts what its text says, and
// what it refuses.

const (
	s82o188Secrets  = "30c4914e-4d9f-453d-bf5b-48ab8e0f2ba3"
	s82o188Lurcher  = "85e18895-d1d2-46c7-afd0-31da410c0438"
	s82o188Molder   = "a6fa840a-817c-4e84-9dd3-d28a011f2214"
	s82o188Raiders  = "abd12238-ea9d-4124-8229-3a479b8464d7"
	s82o188Wurm     = "be843255-7408-45b9-b0ca-c79342daa353"
	s82o188Hatchery = "d1950094-2cc9-48ce-ae1a-0ae78ba252c1"
)

func TestS82Old188EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82o188Secrets:  "Mausoleum Secrets",
		s82o188Lurcher:  "Rhizome Lurcher",
		s82o188Molder:   "Molderhulk",
		s82o188Raiders:  "Golgari Raiders",
		s82o188Wurm:     "Vigorspore Wurm",
		s82o188Hatchery: "Hatchery Spider",
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

// s82o188Creatures puts n creature cards in a player's graveyard.
func s82o188Creatures(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: p.ID, Controller: p.ID})
	}
}

func s82o188Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCardByID(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[game.CounterPlusOne]
}

func TestS82Old188MausoleumSecretsFindsABlackCardUpToTheCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	s82o188Creatures(me, 2)
	cheap := uuid.New()
	pricey := uuid.New()
	green := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: pricey, Name: "Pricey", TypeLine: "Creature — Horror", ManaCost: "{3}{B}", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: green, Name: "Green", TypeLine: "Creature — Elf", ManaCost: "{G}", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: cheap, Name: "Cheap", TypeLine: "Creature — Imp", ManaCost: "{1}{B}", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Mausoleum Secrets", "Instant", s82o188Secrets, nil)
	effectstest.PassPriorityAroundTable(t, g)
	open := effectstest.ChkB1OpenSearch(g)
	if open == nil {
		t.Fatal("no search opened")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range open.SearchCards {
		offered[id] = true
	}
	if !offered[cheap] || offered[pricey] || offered[green] {
		t.Errorf("offered cheap=%v pricey=%v green=%v, want only the cheap black card", offered[cheap], offered[pricey], offered[green])
	}
	if err := g.ResolveSearchLibrary(open.ID, open.Chooser, []uuid.UUID{cheap}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) {
		t.Error("the black card did not reach the hand")
	}
}

func TestS82Old188RhizomeLurcherAndGolgariRaidersCountTheGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Rhizome Lurcher", "Creature — Fungus Zombie", s82o188Lurcher, 2, 2},
		{"Golgari Raiders", "Creature — Elf Warrior", s82o188Raiders, 0, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			s82o188Creatures(me, 3)
			g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Bear"})
			me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
			id := effectstest.CastAndResolveCreature(t, g, tc.name, tc.typeLine, tc.oracle)
			effectstest.PassPriorityAroundTable(t, g)
			if n := s82o188Counters(t, g, id); n != 3 {
				t.Errorf("%s entered with %d counters, want 3 (your creature cards only)", tc.name, n)
			}
		})
	}
}

func TestS82Old188GolgariRaidersWithAnEmptyGraveyardDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Golgari Raiders", TypeLine: "Creature — Elf Warrior", OracleID: s82o188Raiders, ManaCost: "{3}{G}", PrintedPTKnown: true, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(id) {
		t.Error("a 0/0 with no counters stayed on the battlefield")
	}
}

func TestS82Old188MolderhulkCostsLessAndReturnsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	s82o188Creatures(me, 4)
	land := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.Acrb1AddMana(t, g, me.ID, "{B}{G}{B}{B}{B}{B}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Molderhulk", TypeLine: "Creature — Fungus Zombie", OracleID: s82o188Molder, ManaCost: "{7}{B}{G}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	// {7}{B}{G} less four is {3}{B}{G}: five mana from the pool is enough.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell with four creature cards in the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, land)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land card did not return to the battlefield")
	}
}

func TestS82Old188MolderhulkIsNotCheapWithAnEmptyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.Acrb1AddMana(t, g, me.ID, "{B}{G}{B}{B}{B}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Molderhulk", TypeLine: "Creature — Fungus Zombie", OracleID: s82o188Molder, ManaCost: "{7}{B}{G}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Molderhulk was cast for five mana with no creature card in the graveyard")
	}
}

func TestS82Old188VigorsporeWurmPumpsByTheGraveyardAndLimitsBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	s82o188Creatures(me, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wurm := effectstest.CastAndResolveCreature(t, g, "Vigorspore Wurm", "Creature — Wurm", s82o188Wurm)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 5 {
		t.Errorf("the target has power %d, want 5 (+3/+3)", p)
	}
	if !effects.EotHasAbility(effectstest.EffectiveAbilities(t, g, bear), "vigilance") {
		t.Error("the target did not gain vigilance")
	}
	spec, _ := effects.Lookup(s82o188Wurm)
	if len(spec.BlockRules) != 1 {
		t.Errorf("%d block rules, want the one-blocker limit", len(spec.BlockRules))
	}
	_ = wurm
}

func TestS82Old188HatcherySpiderRevealsXAndMayPutAGreenPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	s82o188Creatures(me, 3)
	green := uuid.New()
	redTooBig := uuid.New()
	greenTooBig := uuid.New()
	deep := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: deep, Name: "Deep", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: greenTooBig, Name: "Huge Elf", TypeLine: "Creature — Elf", ManaCost: "{4}{G}", Colors: []string{"G"}, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: redTooBig, Name: "Red Goblin", TypeLine: "Creature — Goblin", ManaCost: "{R}", Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: green, Name: "Small Elf", TypeLine: "Creature — Elf", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.Acrb1AddMana(t, g, me.ID, "{G}{G}{G}{G}{G}{G}{G}")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Hatchery Spider", TypeLine: "Creature — Spider", OracleID: s82o188Hatchery, ManaCost: "{5}{G}{G}", Power: 5, Toughness: 7, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick prompt for the revealed cards")
	}
	offered := map[uuid.UUID]bool{}
	for _, c := range pick.ChooseCards {
		offered[c] = true
	}
	if !offered[green] || offered[redTooBig] || offered[greenTooBig] {
		t.Errorf("offered small green=%v red=%v huge green=%v; only the small green card fits", offered[green], offered[redTooBig], offered[greenTooBig])
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{green}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(green) {
		t.Error("the green permanent did not enter")
	}
	if me.Library.Contains(redTooBig) == false || me.Library.Contains(greenTooBig) == false {
		t.Error("the rest were not kept in the library (bottom, random order)")
	}
}
