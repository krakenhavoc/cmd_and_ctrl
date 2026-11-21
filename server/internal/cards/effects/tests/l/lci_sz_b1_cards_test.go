package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_sz_b1_cards_test.go — The Lost Caverns of Ixalan, S–Z batch 1
// (#4470). Each test plays the real catalog entry and asserts what its
// text says.

const (
	lciSzSurveyorOracle  = "63c8f510-eba8-4f00-9f80-fbbc74801a4a"
	lciSzStalkerOracle   = "c8c28807-8abf-4f91-81f3-1d7d9e856c25"
	lciSzMilitiaOracle   = "19afaf35-b420-49e0-bddf-3180bf1dc30f"
	lciSzVanguardOracle  = "de4678f8-a441-4e4a-a339-9b62ebfb0100"
	lciSzWaterwindOracle = "db65f874-6bf3-4fdb-a977-10b01b7ee2e9"
	lciSzPiratesOracle   = "5453e4fa-49ea-43fc-8de5-6873d29ef627"
)

func TestLciSzB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciSzSurveyorOracle:  "Scampering Surveyor",
		lciSzStalkerOracle:   "Stalactite Stalker",
		lciSzMilitiaOracle:   "Sunshot Militia",
		lciSzVanguardOracle:  "Vanguard of the Rose",
		lciSzWaterwindOracle: "Waterwind Scout",
		lciSzPiratesOracle:   "Waylaying Pirates",
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

// --- Scampering Surveyor -----------------------------------------------------------

func TestLciSzB1ScamperingSurveyorFetchesABasicOrCaveTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Echoing Deeps", "Land — Cave"),
		effectstest.SearchTestLand("Scoured Barrens", "Land — Plains Swamp"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Scampering Surveyor", TypeLine: "Artifact Creature — Gnome", OracleID: lciSzSurveyorOracle, Power: 3, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt when the Surveyor entered")
	}
	if len(c.SearchCards) != 2 {
		t.Fatalf("search offers %d cards, want the basic land and the Cave", len(c.SearchCards))
	}
	for _, id := range c.SearchCards {
		if id == ids[2] || id == ids[3] {
			t.Error("a nonbasic non-Cave card is offered")
		}
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !g.Battlefield.Contains(ids[1]) {
		t.Fatal("the Cave did not reach the battlefield")
	}
	if !effectstest.TappedNow(g, ids[1]) {
		t.Error("the Cave entered untapped")
	}
}

// --- Stalactite Stalker ------------------------------------------------------------

func TestLciSzB1StalactiteStalkerGrowsAtYourEndStepIfYouDescended(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		want     int
	}{
		{"a permanent card milled", "Creature — Bear", 1},
		{"only an instant card milled", "Instant", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			stalker := effectstest.B12Push(g, me.ID, "Stalactite Stalker", "Creature — Goblin Rogue", lciSzStalkerOracle, 1, 1)
			if !effectstest.HasEffectiveKeyword(t, g, stalker, "menace") {
				t.Error("Stalactite Stalker has no menace")
			}
			effectstest.LciDGMillOne(t, g, me, tc.typeLine)
			effectstest.AdvanceToEndStepOf(t, g, 0)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, stalker, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

func TestLciSzB1StalactiteStalkerSacrificesForMinusItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stalker := effectstest.B12Push(g, me.ID, "Stalactite Stalker", "Creature — Goblin Rogue", lciSzStalkerOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(stalker, game.CounterPlusOne, 2) })
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.RfCardTarget(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(stalker) {
		t.Error("the Stalker was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The Stalker was a 3/3 (two counters) as it was sacrificed.
	if p, tough := effectstest.EffectivePower(t, g, victim), effectstest.EffectiveToughness(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1 (a 4/4 gets -3/-3)", p, tough)
	}
}

// --- Sunshot Militia ---------------------------------------------------------------

func TestLciSzB1SunshotMilitiaTapsTwoToPingEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	militia := effectstest.B12Push(g, me.ID, "Sunshot Militia", "Creature — Human Soldier", lciSzMilitiaOracle, 1, 3)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	land := effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, militia, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{militia, land}, Strict: true}); err == nil {
		t.Fatal("a land paid for \"artifacts and/or creatures\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, militia, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{militia, relic}, Strict: true}); err != nil {
		t.Fatalf("tapping the Militia and an artifact: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 1
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("%s at %d, want %d", p.Name, p.Life, want)
		}
	}
}

func TestLciSzB1SunshotMilitiaIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	militia := effectstest.B12Push(g, me.ID, "Sunshot Militia", "Creature — Human Soldier", lciSzMilitiaOracle, 1, 3)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, militia, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{militia, relic}, Strict: true}); err == nil {
		t.Error("the ability was activated at the beginning of combat")
	}
}

// --- Vanguard of the Rose ----------------------------------------------------------

func TestLciSzB1VanguardOfTheRoseGainsIndestructibleAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vanguard := effectstest.B12Push(g, me.ID, "Vanguard of the Rose", "Creature — Vampire Knight", lciSzVanguardOracle, 3, 1)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "My Land", "Basic Land — Plains")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, vanguard, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{vanguard}}); err == nil {
		t.Fatal("the Vanguard paid with itself; the cost is another creature or artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, vanguard, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("a land paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, vanguard, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate with a creature: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, vanguard, "indestructible") {
		t.Error("the Vanguard did not gain indestructible")
	}
	if !effectstest.TappedNow(g, vanguard) {
		t.Error("the Vanguard did not tap")
	}
}

// --- Waterwind Scout ---------------------------------------------------------------

func TestLciSzB1WaterwindScoutMakesAMapOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.Msc17Play(t, g, me, game.Card{Name: "Waterwind Scout", TypeLine: "Creature — Merfolk Scout", OracleID: lciSzWaterwindOracle, Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := lciB1MapsControlledBy(g, me.ID); got != 1 {
		t.Errorf("controller has %d Map tokens, want 1", got)
	}
	for _, p := range g.Seats[1:] {
		if got := lciB1MapsControlledBy(g, p.ID); got != 0 {
			t.Errorf("%s has %d Map tokens, want 0", p.Name, got)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Waterwind Scout has no flying")
	}
}

// --- Waylaying Pirates -------------------------------------------------------------

func TestLciSzB1WaylayingPiratesTapsAndStunsWithAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := effectstest.Msc17Cast(t, g, me, game.Card{Name: "Waylaying Pirates", TypeLine: "Creature — Human Pirate", OracleID: lciSzPiratesOracle, Power: 3, Toughness: 3}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Waylaying Pirates: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target pick for the enters trigger")
	}
	for _, c := range p.PickTargetCards {
		if card, ok := g.LookupCardForEffect(c); ok && card.Controller == me.ID {
			t.Errorf("the controller's own %s is offered", card.Name)
		}
	}
	effectstest.PickIfAsked(t, g, me.ID, theirRock)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, theirRock) {
		t.Error("the target artifact was not tapped")
	}
	if got := effectstest.CountersOn(g, theirRock, game.CounterStun); got != 1 {
		t.Errorf("the target has %d stun counters, want 1", got)
	}
}

func TestLciSzB1WaylayingPiratesDoNothingWithoutAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Msc17Play(t, g, me, game.Card{Name: "Waylaying Pirates", TypeLine: "Creature — Human Pirate", OracleID: lciSzPiratesOracle, Power: 3, Toughness: 3})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, bear) || effectstest.CountersOn(g, bear, game.CounterStun) != 0 {
		t.Error("the trigger resolved though its controller controls no artifact")
	}
}
