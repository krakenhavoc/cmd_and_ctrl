package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b28_cards_test.go — Star Trek, batch 28 (#4973). Each test plays the
// real catalog entry through the engine.

const (
	trekB28EnterpriseDOracle  = "d95af032-3efd-40c7-8229-ade9d974934f"
	trekB28WarpCoreOracle     = "6a2df97a-341d-4bc4-9f5c-661b6ca78fd0"
	trekB28AndoriaOracle      = "79134f5e-94db-4a40-ba75-2569b839cf35"
	trekB28ArtificialOracle   = "7e550295-eaf4-4962-bc52-261d7c2d71ae"
	trekB28BajorOracle        = "93d7f6e7-b6b0-48f3-891a-73970826af3b"
	trekB28CardassiaOracle    = "aff720b4-b8ef-46ea-9a3d-e101999e6526"
	trekB28EarthOracle        = "c75d87b0-7de9-4c50-b781-9749d67c1e07"
	trekB28OutpostOracle      = "a8bcc930-3efe-4595-b164-f5985e1dc289"
	trekB28FerenginarOracle   = "2e0753ad-7f76-4d26-8e87-6f201378f333"
	trekB28GaseousOracle      = "1fb736bf-c9bd-45ab-86ce-7f138f1724b2"
	trekB28EnterpriseDName    = "U.S.S. Enterprise-D, Galaxy-Class"
	trekB28SpacecraftTypeLine = "Legendary Artifact — Spacecraft"
)

func TestTrekB28EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB28EnterpriseDOracle: trekB28EnterpriseDName,
		trekB28WarpCoreOracle:    "Warp Core",
		trekB28AndoriaOracle:     "Andoria",
		trekB28ArtificialOracle:  "Artificial Anomaly",
		trekB28BajorOracle:       "Bajor",
		trekB28CardassiaOracle:   "Cardassia Prime",
		trekB28EarthOracle:       "Earth, Sector 001",
		trekB28OutpostOracle:     "Federation Outpost",
		trekB28FerenginarOracle:  "Ferenginar",
		trekB28GaseousOracle:     "Gaseous Anomaly",
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

// trekB28Land puts an untapped basic land on the battlefield for p.
func trekB28Land(g *game.Game, p *game.Player, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Basic Land — " + name,
		Owner: p.ID, Controller: p.ID,
	})
}

// --- the five Planets ---------------------------------------------------

func TestTrekB28PlanetsEnterTappedAndGainOneLife(t *testing.T) {
	for _, tc := range []struct{ name, oracle, a, b string }{
		{"Andoria", trekB28AndoriaOracle, "U", "R"},
		{"Bajor", trekB28BajorOracle, "R", "W"},
		{"Cardassia Prime", trekB28CardassiaOracle, "W", "B"},
		{"Earth, Sector 001", trekB28EarthOracle, "G", "U"},
		{"Ferenginar", trekB28FerenginarOracle, "B", "G"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		before := me.Life
		id := playLandFromHand(t, g, tc.name, tc.oracle)
		top100AssertEnteredTapped(t, g, id, tc.name)
		if triggerOnStack(g, id) == nil {
			t.Errorf("%s: the life is a trigger with a response window", tc.name)
		}
		passPriorityAroundTable(t, g)
		if me.Life != before+1 {
			t.Errorf("%s: life %d → %d, want +1", tc.name, before, me.Life)
		}
		spec, _ := Lookup(tc.oracle)
		if want := "{" + tc.a + "|" + tc.b + "}"; len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, want)
		}
	}
}

// --- Federation Outpost -------------------------------------------------

func TestTrekB28OutpostEntersTappedWithFewLandsAndTriggersNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trekB28Land(g, me, "Plains")
	id := playLandFromHand(t, g, "Federation Outpost", trekB28OutpostOracle)
	top100AssertEnteredTapped(t, g, id, "Federation Outpost")
	g.RunStateChecksForTest()
	if triggerOnStack(g, id) != nil || modePickChoiceFor(g, me.ID) != nil {
		t.Error("a tapped Outpost has no enters-untapped trigger")
	}
}

func TestTrekB28OutpostEntersUntappedAndGainsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trekB28Land(g, me, "Plains")
	trekB28Land(g, me, "Island")
	before := me.Life
	id := playLandFromHand(t, g, "Federation Outpost", trekB28OutpostOracle)
	g.RunStateChecksForTest()
	if c, ok := battlefieldCard(g, id); !ok || c.Tapped {
		t.Fatalf("with two other lands the Outpost enters untapped (on battlefield %v)", ok)
	}
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick for the enters-untapped trigger: %+v", g.PendingChoices)
	}
	answerMode(t, g, c, me.ID, 0)
	passPriorityAroundTable(t, g)
	if me.Life != before+2 {
		t.Errorf("life %d → %d, want +2", before, me.Life)
	}
}

func TestTrekB28OutpostEntersUntappedAndSurveilsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B")
	trekB28Land(g, me, "Plains")
	trekB28Land(g, me, "Island")
	before := me.Life
	playLandFromHand(t, g, "Federation Outpost", trekB28OutpostOracle)
	g.RunStateChecksForTest()
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no mode_pick for the enters-untapped trigger: %+v", g.PendingChoices)
	}
	answerMode(t, g, c, me.ID, 1)
	passPriorityAroundTable(t, g)
	if n := trekB6SurveilAll(t, g, me.ID); n != 1 {
		t.Errorf("surveil looked at %d cards, want 1", n)
	}
	if me.Life != before {
		t.Errorf("the surveil mode changed life %d → %d", before, me.Life)
	}
}

// --- Artificial Anomaly and Gaseous Anomaly ---------------------------

func trekB28AnomalyBoard(t *testing.T, name, oracle string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Land", OracleID: oracle,
		Owner: me.ID, Controller: me.ID,
	})
	return g, me, id
}

// trekB28AssertTappedNewPlanet checks one New Planet came out tapped
// under p with the printed any-colour mana ability.
func trekB28AssertTappedNewPlanet(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name != "New Planet" {
			continue
		}
		n++
		if !c.Tapped || c.Controller != p.ID || !c.IsLand() || !c.HasSubtype("Planet") {
			t.Errorf("New Planet = tapped %v, mine %v, land %v, Planet %v", c.Tapped, c.Controller == p.ID, c.IsLand(), c.HasSubtype("Planet"))
		}
	}
	if n != 1 {
		t.Errorf("%d New Planet tokens, want 1", n)
	}
}

func TestTrekB28ArtificialAnomalyDrainsTwoAndMakesANewPlanet(t *testing.T) {
	g, me, land := trekB28AnomalyBoard(t, "Artificial Anomaly", trekB28ArtificialOracle)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	floatMana(t, g, me, "{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the land is sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 2
		if p.ID == me.ID {
			want = lives[p.ID] + 2
		}
		if p.Life != want {
			t.Errorf("seat %s life %d → %d, want %d", p.Name, lives[p.ID], p.Life, want)
		}
	}
	trekB28AssertTappedNewPlanet(t, g, me)
}

func TestTrekB28ArtificialAnomalyNeedsItsMana(t *testing.T) {
	g, me, land := trekB28AnomalyBoard(t, "Artificial Anomaly", trekB28ArtificialOracle)
	floatMana(t, g, me, "{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with {3}{U} for a {3}{B} cost")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("a refused activation sacrificed the land")
	}
}

func TestTrekB28GaseousAnomalyDrawsTwoAndMakesANewPlanet(t *testing.T) {
	g, me, land := trekB28AnomalyBoard(t, "Gaseous Anomaly", trekB28GaseousOracle)
	seedLibrary(me, "A", "B", "C")
	hand := me.Hand.Size()
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2", hand, me.Hand.Size())
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	trekB28AssertTappedNewPlanet(t, g, me)
}

func TestTrekB28GaseousAnomalyNeedsItsMana(t *testing.T) {
	g, me, land := trekB28AnomalyBoard(t, "Gaseous Anomaly", trekB28GaseousOracle)
	floatMana(t, g, me, "{C}{C}{C}{C}{U}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with {4}{U} for a {5}{U} cost")
	}
}

// --- Warp Core -----------------------------------------------------------

func TestTrekB28WarpCoreFetchesABasicLandToHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Volcanic Island", "Land — Island Mountain"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	castCatalogSpell(t, g, "Warp Core", "Artifact", trekB28WarpCoreOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the enters trigger asked for no search")
	}
	if searchOptionNamed(g, c, "Volcanic Island") != uuid.Nil || searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a nonbasic land or a nonland card is offered")
	}
	hand := me.Hand.Size()
	answerSearchNamed(t, g, me.ID, "Island")
	if me.Hand.Size() != hand+1 || !b02bHandHasNamed(me, "Island") {
		t.Errorf("hand %d → %d: want the basic Island in hand", hand, me.Hand.Size())
	}
}

func TestTrekB28WarpCoreMayFailToFind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, searchTestLand("Island", "Basic Land — Island"))
	castCatalogSpell(t, g, "Warp Core", "Artifact", trekB28WarpCoreOracle, nil)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerSearchFailToFind(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d after declining to find", hand, me.Hand.Size())
	}
}

func trekB28WarpCoreBoard(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	core := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Warp Core", TypeLine: "Artifact", OracleID: trekB28WarpCoreOracle,
		Owner: me.ID, Controller: me.ID,
	})
	return g, me, core
}

func trekB28ChargeTarget(t *testing.T, g *game.Game, me *game.Player, core, target uuid.UUID) error {
	t.Helper()
	floatMana(t, g, me, "{C}")
	return g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
}

func TestTrekB28WarpCorePutsTwoChargeCountersOnASpacecraftAsASorcery(t *testing.T) {
	g, me, core := trekB28WarpCoreBoard(t)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Ship", TypeLine: "Artifact — Spacecraft",
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID,
	})
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plain Rock", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})

	// The harness starts outside a main phase: not a sorcery window.
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("activated outside a main phase")
	}
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Fatal("a non-Spacecraft artifact was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !layeredCard(t, g, core).Tapped {
		t.Error("{T} is part of the cost")
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, theirs, game.CounterCharge); got != 2 {
		t.Errorf("charge counters on the opponent's Spacecraft = %d, want 2", got)
	}
}

func TestTrekB28WarpCoreNeedsItsMana(t *testing.T) {
	g, me, core := trekB28WarpCoreBoard(t)
	ship := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Ship", TypeLine: "Artifact — Spacecraft",
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ship}},
	}); err == nil {
		t.Fatal("activated with no mana for a {1} cost")
	}
}

// --- U.S.S. Enterprise-D, Galaxy-Class ----------------------------------

func TestTrekB28EnterpriseDBecomesAFlyingVigilantFourFiveAtSeven(t *testing.T) {
	g, me, ship, crew := stationBoard(t, trekB28EnterpriseDName, trekB28SpacecraftTypeLine, trekB28EnterpriseDOracle, 6)
	seedLibrary(me, "X", "Y")
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, ship); c.IsCreature() || game.HasKeyword(&c, "flying") {
		t.Error("six charge counters is below 7+")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(ship, game.CounterCharge, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	at7 := layeredCard(t, g, ship)
	if !at7.IsCreature() || !game.HasKeyword(&at7, "flying") || !game.HasKeyword(&at7, "vigilance") ||
		at7.CurrentPower() != 4 || at7.CurrentToughness() != 5 {
		t.Errorf("at 7+: creature %v flying %v vigilance %v %d/%d, want a 4/5 flier with vigilance",
			at7.IsCreature(), game.HasKeyword(&at7, "flying"), game.HasKeyword(&at7, "vigilance"),
			at7.CurrentPower(), at7.CurrentToughness())
	}
}

func TestTrekB28EnterpriseDImpulsesOnTheFirstChargeCountersEachTurn(t *testing.T) {
	g, me, ship, crew := stationBoard(t, trekB28EnterpriseDName, trekB28SpacecraftTypeLine, trekB28EnterpriseDOracle, 2)
	top := seedLibrary(me, "Top Card", "Second Card")
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ship, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters = %d, want 2", got)
	}
	if !g.Exile.Contains(top[0]) {
		t.Fatal("the first charge counters this turn did not exile the top card")
	}
	perm := exiledPermission(g, top[0])
	if perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
		t.Errorf("permission %+v: want a live PLAY grant for the controller", perm)
	}

	// A second placement the same turn — here Warp Core's — is not the first.
	core := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Warp Core", TypeLine: "Artifact", OracleID: trekB28WarpCoreOracle,
		Owner: me.ID, Controller: me.ID,
	})
	if err := trekB28ChargeTarget(t, g, me, core, ship); err != nil {
		t.Fatalf("Warp Core: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ship, game.CounterCharge); got != 4 {
		t.Fatalf("charge counters = %d, want 4", got)
	}
	if g.Exile.Contains(top[1]) {
		t.Error("a second placement in one turn exiled another card")
	}
}

func TestTrekB28EnterpriseDTriggersAgainNextTurnFromAnotherSource(t *testing.T) {
	g, me, ship, crew := stationBoard(t, trekB28EnterpriseDName, trekB28SpacecraftTypeLine, trekB28EnterpriseDOracle, 2)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	seat := g.Turn.ActiveSeat
	advanceToStepOf(t, g, (seat+1)%len(g.Seats), game.StepUpkeep)
	advanceToStepOf(t, g, seat, game.StepPrecombatMain)
	top := seedLibrary(me, "Next Turn Card")[0]
	core := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Warp Core", TypeLine: "Artifact", OracleID: trekB28WarpCoreOracle,
		Owner: me.ID, Controller: me.ID,
	})
	if err := trekB28ChargeTarget(t, g, me, core, ship); err != nil {
		t.Fatalf("Warp Core: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Error("the first charge counters of a new turn, from Warp Core, did not exile the top card")
	}
}

func TestTrekB28EnterpriseDIgnoresOtherCountersAndRemovals(t *testing.T) {
	g, me, ship, _ := stationBoard(t, trekB28EnterpriseDName, trekB28SpacecraftTypeLine, trekB28EnterpriseDOracle, 2)
	top := seedLibrary(me, "Top Card")[0]
	other := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Other Ship", TypeLine: "Artifact — Spacecraft",
		Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(ship, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
		if err := g.AddCounterForEffect(other, game.CounterCharge, 3); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(top) {
		t.Fatal("a +1/+1 counter, or charge counters on another Spacecraft, triggered Enterprise-D")
	}
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(ship, game.CounterCharge, 2); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
		if err := g.AddCounterForEffect(ship, game.CounterCharge, -1); err != nil {
			t.Fatalf("AddCounterForEffect: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("charge counters put on Enterprise-D did not trigger it")
	}
	if n := trekB28ExiledCount(g, me); n != 1 {
		t.Errorf("%d cards exiled, want 1 — a removal is not a placement", n)
	}
}

func trekB28ExiledCount(g *game.Game, p *game.Player) int {
	n := 0
	for _, c := range g.Exile.Cards {
		if c.Owner == p.ID {
			n++
		}
	}
	return n
}
