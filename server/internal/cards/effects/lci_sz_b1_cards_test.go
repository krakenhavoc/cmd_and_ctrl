package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Scampering Surveyor -----------------------------------------------------------

func TestLciSzB1ScamperingSurveyorFetchesABasicOrCaveTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Echoing Deeps", "Land — Cave"),
		searchTestLand("Scoured Barrens", "Land — Plains Swamp"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	msc17Play(t, g, me, game.Card{Name: "Scampering Surveyor", TypeLine: "Artifact Creature — Gnome", OracleID: lciSzSurveyorOracle, Power: 3, Toughness: 2})
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
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
	if !tappedNow(g, ids[1]) {
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			stalker := b12Push(g, me.ID, "Stalactite Stalker", "Creature — Goblin Rogue", lciSzStalkerOracle, 1, 1)
			if !hasEffectiveKeyword(t, g, stalker, "menace") {
				t.Error("Stalactite Stalker has no menace")
			}
			lciDGMillOne(t, g, me, tc.typeLine)
			advanceToEndStepOf(t, g, 0)
			passPriorityAroundTable(t, g)
			if got := countersOn(g, stalker, game.CounterPlusOne); got != tc.want {
				t.Errorf("%d +1/+1 counters, want %d", got, tc.want)
			}
		})
	}
}

func TestLciSzB1StalactiteStalkerSacrificesForMinusItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stalker := b12Push(g, me.ID, "Stalactite Stalker", "Creature — Goblin Rogue", lciSzStalkerOracle, 1, 1)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(stalker, game.CounterPlusOne, 2) })
	victim := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	advanceToMain(t, g)
	fillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, stalker, 0, game.ActivateAbilityParams{Strict: true, Targets: rfCardTarget(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(stalker) {
		t.Error("the Stalker was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	// The Stalker was a 3/3 (two counters) as it was sacrificed.
	if p, tough := effectivePower(t, g, victim), effectiveToughness(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("the target is %d/%d, want 1/1 (a 4/4 gets -3/-3)", p, tough)
	}
}

// --- Sunshot Militia ---------------------------------------------------------------

func TestLciSzB1SunshotMilitiaTapsTwoToPingEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	militia := b12Push(g, me.ID, "Sunshot Militia", "Creature — Human Soldier", lciSzMilitiaOracle, 1, 3)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	land := pushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
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
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	militia := b12Push(g, me.ID, "Sunshot Militia", "Creature — Human Soldier", lciSzMilitiaOracle, 1, 3)
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, militia, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{militia, relic}, Strict: true}); err == nil {
		t.Error("the ability was activated at the beginning of combat")
	}
}

// --- Vanguard of the Rose ----------------------------------------------------------

func TestLciSzB1VanguardOfTheRoseGainsIndestructibleAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vanguard := b12Push(g, me.ID, "Vanguard of the Rose", "Creature — Vampire Knight", lciSzVanguardOracle, 3, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "My Land", "Basic Land — Plains")
	advanceToMain(t, g)
	fillPoolColored(me, "W", 1)
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
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, vanguard, "indestructible") {
		t.Error("the Vanguard did not gain indestructible")
	}
	if !tappedNow(g, vanguard) {
		t.Error("the Vanguard did not tap")
	}
}

// --- Waterwind Scout ---------------------------------------------------------------

func TestLciSzB1WaterwindScoutMakesAMapOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := msc17Play(t, g, me, game.Card{Name: "Waterwind Scout", TypeLine: "Creature — Merfolk Scout", OracleID: lciSzWaterwindOracle, Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if got := lciB1MapsControlledBy(g, me.ID); got != 1 {
		t.Errorf("controller has %d Map tokens, want 1", got)
	}
	for _, p := range g.Seats[1:] {
		if got := lciB1MapsControlledBy(g, p.ID); got != 0 {
			t.Errorf("%s has %d Map tokens, want 0", p.Name, got)
		}
	}
	if !hasEffectiveKeyword(t, g, id, "flying") {
		t.Error("Waterwind Scout has no flying")
	}
}

// --- Waylaying Pirates -------------------------------------------------------------

func TestLciSzB1WaylayingPiratesTapsAndStunsWithAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirRock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if _, err := msc17Cast(t, g, me, game.Card{Name: "Waylaying Pirates", TypeLine: "Creature — Human Pirate", OracleID: lciSzPiratesOracle, Power: 3, Toughness: 3}, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Waylaying Pirates: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target pick for the enters trigger")
	}
	for _, c := range p.PickTargetCards {
		if card, ok := g.LookupCardForEffect(c); ok && card.Controller == me.ID {
			t.Errorf("the controller's own %s is offered", card.Name)
		}
	}
	pickIfAsked(t, g, me.ID, theirRock)
	passPriorityAroundTable(t, g)
	if !tappedNow(g, theirRock) {
		t.Error("the target artifact was not tapped")
	}
	if got := countersOn(g, theirRock, game.CounterStun); got != 1 {
		t.Errorf("the target has %d stun counters, want 1", got)
	}
}

func TestLciSzB1WaylayingPiratesDoNothingWithoutAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	msc17Play(t, g, me, game.Card{Name: "Waylaying Pirates", TypeLine: "Creature — Human Pirate", OracleID: lciSzPiratesOracle, Power: 3, Toughness: 3})
	passPriorityAroundTable(t, g)
	if tappedNow(g, bear) || countersOn(g, bear, game.CounterStun) != 0 {
		t.Error("the trigger resolved though its controller controls no artifact")
	}
}
