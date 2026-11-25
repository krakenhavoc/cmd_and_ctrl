package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander 2020, OLD batch 121: counters.

const (
	oldB121Thalia      = "91508175-fa1d-4c51-99d4-036a2de6abb0"
	oldB121Vastwood    = "7f10540f-eaba-4e06-a79e-101f7409ac6c"
	oldB121Hungering   = "d1dde190-8124-4590-9507-801ea4c8cde3"
	oldB121Lavabrink   = "bd4df73c-bbd5-47fd-a961-19f3a3c59381"
	oldB121Kathril     = "730897a9-3ead-4990-9492-1a2040d8ccac"
	oldB121Tayam       = "84be8354-0bfa-4df1-868a-13197bd49191"
	oldB121PlusOne     = game.CounterPlusOne
	oldB121HumanType   = "Creature — Human Soldier"
	oldB121BearType    = "Creature — Bear"
	oldB121NightmareTL = "Legendary Creature — Nightmare Insect"
)

func TestOldB121CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB121Thalia: "Thalia's Lieutenant", oldB121Vastwood: "Vastwood Hydra", oldB121Hungering: "Hungering Hydra",
		oldB121Lavabrink: "Lavabrink Floodgates", oldB121Kathril: "Kathril, Aspect Warper", oldB121Tayam: "Tayam, Luminous Enigma",
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

func oldB121Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	return effectstest.AvB1Card(t, g, id).Counters[kind]
}

func TestOldB121ThaliasLieutenantGrowsHumansBothWays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h1 := effectstest.B12Creature(g, me.ID, "Human One", oldB121HumanType, 1, 1)
	h2 := effectstest.B12Creature(g, me.ID, "Human Two", "Creature — Human", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB121BearType, 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Their Human", "Creature — Human", 1, 1)
	lt := effectstest.AvB1ResolveCreature(t, g, me, "Thalia's Lieutenant", oldB121HumanType, oldB121Thalia)
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"first Human", h1, 1}, {"second Human", h2, 1}, {"Bear", bear, 0}, {"their Human", foe, 0}, {"the Lieutenant", lt, 0}} {
		if n := oldB121Counters(t, g, tc.id, oldB121PlusOne); n != tc.want {
			t.Errorf("%s has %d counters, want %d", tc.name, n, tc.want)
		}
	}
	// Another Human entering grows the Lieutenant; a Bear or their Human does not.
	effectstest.AvB1ResolveCreature(t, g, me, "Bear Two", oldB121BearType, "")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AvB1OpponentCreatureEnters(t, g, opp.ID, "Creature — Human")
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB121Counters(t, g, lt, oldB121PlusOne); n != 0 {
		t.Fatalf("a Bear or an opponent's Human grew the Lieutenant: %d", n)
	}
	effectstest.AvB1ResolveCreature(t, g, me, "Human Three", "Creature — Human", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB121Counters(t, g, lt, oldB121PlusOne); n != 1 {
		t.Fatalf("Lieutenant counters %d, want 1", n)
	}
}

func TestOldB121VastwoodHydraDistributesItsCountersWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB121BearType, 2, 2)
	hydra := effectstest.CastXSpell(t, g, "Vastwood Hydra", "Creature — Hydra", oldB121Vastwood, "{X}{G}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB121Counters(t, g, hydra, oldB121PlusOne); n != 3 {
		t.Fatalf("X=3 Hydra entered with %d counters", n)
	}
	effectstest.B14Kill(g, hydra)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hydra) {
		t.Fatal("the Hydra is still on the battlefield")
	}
	if n := oldB121Counters(t, g, bear, oldB121PlusOne); n != 3 {
		t.Errorf("the only other creature has %d counters, want 3", n)
	}
}

func TestOldB121VastwoodHydraAsksWhichCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	b1 := effectstest.B12Creature(g, me.ID, "Bear One", oldB121BearType, 2, 2)
	b2 := effectstest.B12Creature(g, me.ID, "Bear Two", oldB121BearType, 2, 2)
	hydra := effectstest.CastXSpell(t, g, "Vastwood Hydra", "Creature — Hydra", oldB121Vastwood, "{X}{G}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B14Kill(g, hydra)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{b2}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if n1, n2 := oldB121Counters(t, g, b1, oldB121PlusOne), oldB121Counters(t, g, b2, oldB121PlusOne); n1 != 0 || n2 != 2 {
		t.Errorf("counters %d and %d, want 0 and 2", n1, n2)
	}
}

func TestOldB121HungeringHydraGrowsWhenItSurvivesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.B12Creature(g, me.ID, "Pinger", oldB121BearType, 2, 2)
	hydra := effectstest.CastXSpell(t, g, "Hungering Hydra", "Creature — Hydra", oldB121Hungering, "{X}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB121Counters(t, g, hydra, oldB121PlusOne); n != 3 {
		t.Fatalf("X=3 Hydra entered with %d counters", n)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, hydra, 2) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB121Counters(t, g, hydra, oldB121PlusOne); n != 5 {
		t.Errorf("after 2 damage the Hydra has %d counters, want 5", n)
	}
	// Lethal damage kills it before the trigger resolves: no counters.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, hydra, 7) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hydra) {
		t.Error("a Hydra dealt lethal damage survived")
	}
	_ = me
	spec, _ := effects.Lookup(oldB121Hungering)
	if len(spec.BlockRules) != 1 {
		t.Errorf("block rules %d, want the one cap on blockers", len(spec.BlockRules))
	}
}

func oldB121Floodgates(g *game.Game, owner uuid.UUID, doom int) uuid.UUID {
	id := effectstest.ApaPush(g, owner, owner, effectstest.StCard("Lavabrink Floodgates", oldB121Lavabrink, "Artifact", 0, 0))
	if doom > 0 {
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, "doom", doom) })
	}
	return id
}

func oldB121Pick(t *testing.T, g *game.Game, chooser uuid.UUID, want int) {
	t.Helper()
	c := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil {
		t.Fatal("no option prompt is open")
	}
	if c.Chooser != chooser {
		t.Fatalf("asked %v, want %v", c.Chooser, chooser)
	}
	if err := g.ResolveOptionPick(c.ID, chooser, want); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB121LavabrinkFloodgatesTheUpkeepPlayerChooses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fg := oldB121Floodgates(g, me.ID, 0)
	effectstest.BeginUpkeep(t, g, opp.ID)
	c := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if c == nil || len(c.PickOptions) != 2 {
		t.Fatalf("with no doom counter the player is offered %+v, want nothing or put", c)
	}
	oldB121Pick(t, g, opp.ID, 1)
	if n := oldB121Counters(t, g, fg, "doom"); n != 1 {
		t.Fatalf("doom counters %d, want 1", n)
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	if c := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick); c == nil || len(c.PickOptions) != 3 {
		t.Fatalf("with a doom counter the player is offered %+v, want three options", c)
	}
	oldB121Pick(t, g, me.ID, 2)
	if n := oldB121Counters(t, g, fg, "doom"); n != 0 {
		t.Fatalf("doom counters %d after removing, want 0", n)
	}
	effectstest.BeginUpkeep(t, g, me.ID)
	oldB121Pick(t, g, me.ID, 0)
	if n := oldB121Counters(t, g, fg, "doom"); n != 0 {
		t.Fatalf("doing nothing changed the counters to %d", n)
	}
}

func TestOldB121LavabrinkFloodgatesSacrificesAtThreeAndDamagesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fg := oldB121Floodgates(g, me.ID, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB121BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 5)
	effectstest.BeginUpkeep(t, g, opp.ID)
	oldB121Pick(t, g, opp.ID, 1)
	if g.Battlefield.Contains(fg) {
		t.Fatal("the Floodgates survived its third doom counter")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("6 damage left a creature alive")
	}
}

func TestOldB121LavabrinkFloodgatesTapsForRR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fg := oldB121Floodgates(g, me.ID, 0)
	if err := g.ActivateManaAbility(me.ID, fg, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	red := 0
	for _, m := range me.ManaPool {
		if m.Color == "R" {
			red += m.N()
		}
	}
	if red != 2 {
		t.Errorf("pool holds %d red mana, want 2", red)
	}
}

func TestOldB121KathrilPutsKeywordCountersAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushGraveyardTyped(me, "Fallen Bird", "Creature — Bird", "flying", "first strike")
	effectstest.PushGraveyardTyped(me, "Fallen Drake", "Creature — Drake", "flying")
	// A noncreature card with a keyword does not count.
	effectstest.PushGraveyardTyped(me, "Fallen Relic", "Artifact", "lifelink")
	id := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Kathril, Aspect Warper", TypeLine: oldB121NightmareTL, OracleID: oldB121Kathril, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// Kathril is the only creature: flying and first strike go on it, +2/+2.
	for _, kind := range []string{game.CounterFlying, game.CounterFirstStrike} {
		if n := oldB121Counters(t, g, id, kind); n != 1 {
			t.Errorf("%s counters %d, want 1", kind, n)
		}
	}
	if n := oldB121Counters(t, g, id, game.CounterLifelink); n != 0 {
		t.Errorf("a noncreature card's lifelink gave a counter")
	}
	if n := oldB121Counters(t, g, id, oldB121PlusOne); n != 2 {
		t.Errorf("+1/+1 counters %d, want 2", n)
	}
	if !effectstest.HasKeywordOf(g, id, "flying") || !effectstest.HasKeywordOf(g, id, "first strike") {
		t.Error("the counters did not grant the keywords")
	}
}

func TestOldB121KathrilAsksWhereEachCounterGoes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB121BearType, 2, 2)
	effectstest.PushGraveyardTyped(me, "Fallen Bird", "Creature — Bird", "flying", "trample")
	id := effectstest.AvB1Cast(t, g, me, game.Card{Name: "Kathril, Aspect Warper", TypeLine: oldB121NightmareTL, OracleID: oldB121Kathril, Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("flying pick: %v", err)
	}
	c = effectstest.ChooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{id}); err != nil {
		t.Fatalf("trample pick: %v", err)
	}
	if oldB121Counters(t, g, bear, game.CounterFlying) != 1 || oldB121Counters(t, g, id, game.CounterTrample) != 1 {
		t.Error("a keyword counter did not go where it was sent")
	}
	if n := oldB121Counters(t, g, id, oldB121PlusOne); n != 2 {
		t.Errorf("Kathril +1/+1 counters %d, want 2", n)
	}
}

func TestOldB121TayamGivesVigilanceCountersToOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tayam := effectstest.AvB1ResolveCreature(t, g, me, "Tayam, Luminous Enigma", "Legendary Creature — Nightmare Beast", oldB121Tayam)
	if n := oldB121Counters(t, g, tayam, game.CounterVigilance); n != 0 {
		t.Errorf("Tayam gave itself %d vigilance counters", n)
	}
	bear := effectstest.AvB1ResolveCreature(t, g, me, "Bear", oldB121BearType, "")
	if n := oldB121Counters(t, g, bear, game.CounterVigilance); n != 1 {
		t.Errorf("my Bear has %d vigilance counters, want 1", n)
	}
	if !effectstest.HasKeywordOf(g, bear, "vigilance") {
		t.Error("the counter did not grant vigilance")
	}
	effectstest.AvB1OpponentCreatureEnters(t, g, opp.ID, oldB121BearType)
	for _, c := range g.Battlefield.Cards {
		if c.Controller == opp.ID && c.Counters[game.CounterVigilance] != 0 {
			t.Error("an opponent's creature got a vigilance counter")
		}
	}
}

func TestOldB121TayamMillsThenReturnsAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tayam := effectstest.B12Push(g, me.ID, "Tayam, Luminous Enigma", "Legendary Creature — Nightmare Beast", oldB121Tayam, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Counter Bear", oldB121BearType, 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, oldB121PlusOne, 2) })
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(tayam, "charge", 1) })
	cheap := effectstest.PushGraveyardPermanent(me, "Cheap Relic", "Artifact", "{2}")
	effectstest.PushGraveyardPermanent(me, "Costly Relic", "Artifact", "{5}")
	effectstest.ApaMana(me, "C", "C", "C")
	err := g.ActivateCatalogAbility(me.ID, tayam, 0, game.ActivateAbilityParams{
		Strict:           true,
		CounterSourceIDs: []uuid.UUID{bear, tayam},
		CounterCounts:    []int{2, 1},
		CounterKinds:     []string{oldB121PlusOne, "charge"},
	})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := oldB121Counters(t, g, bear, oldB121PlusOne); n != 0 {
		t.Errorf("the Bear still has %d counters", n)
	}
	if !g.Battlefield.Contains(cheap) {
		t.Error("the one legal permanent card was not returned")
	}
}
