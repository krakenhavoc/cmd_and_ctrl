package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dft_sz_cards_test.go — Aetherdrift (DFT), the first batch of cards from
// S to Z (#4464). Each test plays the real catalog entry and asserts what
// its text says. The fixtures (dftABPush, dftABCast, dftABCrew,
// dftABCardRef) are dft_ab_cards_test.go's.

const (
	dftSZSkystreakOracle   = "2f920bf2-6a49-47c2-be3c-9787c377673b"
	dftSZSpinOutOracle     = "c1b6795a-1ac6-4edb-bcee-4db1f1fbe8ea"
	dftSZSkycrashOracle    = "2343ebff-edf4-4132-ad9e-7800f8014072"
	dftSZSpectralOracle    = "cb5ed146-5079-46ef-a3e0-e584fa4467f6"
	dftSZSkyboxFerryOracle = "09726c3d-f700-4bf4-a9e7-326fe41b0a4a"
	dftSZSpotcycleOracle   = "0f1ada92-538b-40cb-9bf3-5c415f5325e6"
)

func TestDftSZEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dftSZSkystreakOracle:   "Skystreak Engineer",
		dftSZSpinOutOracle:     "Spin Out",
		dftSZSkycrashOracle:    "Skycrash",
		dftSZSpectralOracle:    "Spectral Interference",
		dftSZSkyboxFerryOracle: "Skybox Ferry",
		dftSZSpotcycleOracle:   "Spotcycle Scouter",
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

// dftSZCycles cycles a catalog card from the active seat's hand with the
// given mana floating, through its activated ability at index idx, and
// checks that it was discarded and drew the top card of the library.
func dftSZCycles(t *testing.T, name, typeLine, oracle, mana string, idx int) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "Cycled Into")
	id := pushCatalogHandCard(me, name, typeLine, oracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: cycle: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Errorf("%s was not discarded to cycle", name)
	}
	if !b02bHandHasNamed(me, "Cycled Into") {
		t.Errorf("cycling %s drew nothing", name)
	}
}

// --- Skystreak Engineer -----------------------------------------------------------

func TestDftSZSkystreakEngineerFliesAndExhaustsForTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eng := dftABPush(g, me.ID, "Skystreak Engineer", "Creature — Human Pilot", dftSZSkystreakOracle, 1, 3)
	advanceToMain(t, g)
	if !hasEffectiveKeyword(t, g, eng, "flying") {
		t.Error("Skystreak Engineer has no flying")
	}
	fillPool(me, 4)
	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, eng, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, eng, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	fillPool(me, 4)
	fillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, eng, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
}

// --- Spin Out ---------------------------------------------------------------------

func TestDftSZSpinOutDestroysACreatureOrAnUncrewedVehicle(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := dftABPush(g, opp.ID, "Their Vehicle", dftABVehicle, "", 3, 3)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	advanceToMain(t, g)

	if err := castCatalogSpellErr(t, g, "Spin Out", "Instant", dftSZSpinOutOracle, dftABCardRef(rock)); err == nil {
		t.Fatal("Spin Out targeted a noncreature, non-Vehicle artifact")
	}
	castCatalogSpell(t, g, "Spin Out", "Instant", dftSZSpinOutOracle, dftABCardRef(vehicle))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vehicle) {
		t.Error("the uncrewed Vehicle was not destroyed")
	}
	castCatalogSpell(t, g, "Spin Out", "Instant", dftSZSpinOutOracle, dftABCardRef(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not destroyed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was touched")
	}
}

// --- Skycrash ---------------------------------------------------------------------

func TestDftSZSkycrashDestroysAnArtifactOnly(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := castCatalogSpellErr(t, g, "Skycrash", "Instant", dftSZSkycrashOracle, dftABCardRef(bear)); err == nil {
		t.Fatal("Skycrash targeted a nonartifact creature")
	}
	castCatalogSpell(t, g, "Skycrash", "Instant", dftSZSkycrashOracle, dftABCardRef(rock))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

func TestDftSZSkycrashCyclesForRed(t *testing.T) {
	dftSZCycles(t, "Skycrash", "Instant", dftSZSkycrashOracle, "{R}", 0)
}

// --- Spectral Interference --------------------------------------------------------

func TestDftSZSpectralInterferenceTaxesACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castFromSeat(t, g, them, "Spectral Interference", "Instant", dftSZSpectralOracle, dftABCardRef(victim))
	passUntilTaxed(t, g, me.ID)
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the {4} was unanswered")
	}
	answerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("declining the tax must counter the spell")
	}
}

func TestDftSZSpectralInterferenceRefusesAnInstant(t *testing.T) {
	g := newCatalogGame(t)
	them := g.Seats[1]
	toMainForCost(t, g)
	victim := castCatalogSpell(t, g, "Their Instant", "Instant", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Spectral Interference", TypeLine: "Instant",
		OracleID: dftSZSpectralOracle, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Targets: dftABCardRef(victim)}); err == nil {
		t.Error("Spectral Interference targeted an instant spell")
	}
}

// --- Skybox Ferry -----------------------------------------------------------------

func TestDftSZSkyboxFerryCrewsTwoAndFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ferry := dftABPush(g, me.ID, "Skybox Ferry", dftABVehicle, dftSZSkyboxFerryOracle, 4, 4)
	one := b12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	two := b12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ferry, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Fatal("power 1 crewed a Crew 2 Vehicle")
	}
	dftABCrew(t, g, me, ferry, 0, two)
	if !hasEffectiveKeyword(t, g, ferry, "flying") {
		t.Error("the crewed Ferry has no flying")
	}
	if p, tough := effectivePTOf(t, g, ferry); p != 4 || tough != 4 {
		t.Errorf("the crewed Ferry is %d/%d, want 4/4", p, tough)
	}
}

func TestDftSZSkyboxFerryCyclesForTwo(t *testing.T) {
	dftSZCycles(t, "Skybox Ferry", dftABVehicle, dftSZSkyboxFerryOracle, "{C}{C}", 1)
}

// --- Spotcycle Scouter ------------------------------------------------------------

func TestDftSZSpotcycleScouterScriesTwoAndCrewsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedLibrary(me, "Top", "Second", "Third")
	scouter := dftABCast(t, g, me, game.Card{Name: "Spotcycle Scouter", TypeLine: dftABVehicle,
		OracleID: dftSZSpotcycleOracle, Power: 3, Toughness: 2})
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt after the Scouter entered")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("looked at %d cards, want 2", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := libraryTopNames(me, 1); got[0] != "Third" {
		t.Errorf("library top is %q, want Third after bottoming two", got[0])
	}

	// Crew is not a {T} ability of the Vehicle: the Scouter crews the
	// turn it enters.
	pilot := b12Creature(g, me.ID, "Pilot", "Creature — Human Pilot", 1, 1)
	dftABCrew(t, g, me, scouter, 0, pilot)
}
