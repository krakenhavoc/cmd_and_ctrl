package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// dftSZCycles cycles a catalog card from the active seat's hand with the
// given mana floating, through its activated ability at index idx, and
// checks that it was discarded and drew the top card of the library.
func dftSZCycles(t *testing.T, name, typeLine, oracle, mana string, idx int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "Cycled Into")
	id := effectstest.PushCatalogHandCard(me, name, typeLine, oracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("%s: cycle: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) {
		t.Errorf("%s was not discarded to cycle", name)
	}
	if !effectstest.B02bHandHasNamed(me, "Cycled Into") {
		t.Errorf("cycling %s drew nothing", name)
	}
}

// --- Skystreak Engineer -----------------------------------------------------------

func TestDftSZSkystreakEngineerFliesAndExhaustsForTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eng := effectstest.DftABPush(g, me.ID, "Skystreak Engineer", "Creature — Human Pilot", dftSZSkystreakOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, eng, "flying") {
		t.Error("Skystreak Engineer has no flying")
	}
	effectstest.FillPool(me, 4)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, eng, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("exhaust: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, eng, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	effectstest.FillPool(me, 4)
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.ActivateCatalogAbility(me.ID, eng, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrAbilityExhausted) {
		t.Fatalf("second activation: %v, want ErrAbilityExhausted", err)
	}
}

// --- Spin Out ---------------------------------------------------------------------

func TestDftSZSpinOutDestroysACreatureOrAnUncrewedVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	vehicle := effectstest.DftABPush(g, opp.ID, "Their Vehicle", effectstest.DftABVehicle, "", 3, 3)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	effectstest.AdvanceToMain(t, g)

	if err := effectstest.CastCatalogSpellErr(t, g, "Spin Out", "Instant", dftSZSpinOutOracle, effectstest.DftABCardRef(rock)); err == nil {
		t.Fatal("Spin Out targeted a noncreature, non-Vehicle artifact")
	}
	effectstest.CastCatalogSpell(t, g, "Spin Out", "Instant", dftSZSpinOutOracle, effectstest.DftABCardRef(vehicle))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vehicle) {
		t.Error("the uncrewed Vehicle was not destroyed")
	}
	effectstest.CastCatalogSpell(t, g, "Spin Out", "Instant", dftSZSpinOutOracle, effectstest.DftABCardRef(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not destroyed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was touched")
	}
}

// --- Skycrash ---------------------------------------------------------------------

func TestDftSZSkycrashDestroysAnArtifactOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, opp.ID, "Mind Stone", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Skycrash", "Instant", dftSZSkycrashOracle, effectstest.DftABCardRef(bear)); err == nil {
		t.Fatal("Skycrash targeted a nonartifact creature")
	}
	effectstest.CastCatalogSpell(t, g, "Skycrash", "Instant", dftSZSkycrashOracle, effectstest.DftABCardRef(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

func TestDftSZSkycrashCyclesForRed(t *testing.T) {
	dftSZCycles(t, "Skycrash", "Instant", dftSZSkycrashOracle, "{R}", 0)
}

// --- Spectral Interference --------------------------------------------------------

func TestDftSZSpectralInterferenceTaxesACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastFromSeat(t, g, them, "Spectral Interference", "Instant", dftSZSpectralOracle, effectstest.DftABCardRef(victim))
	effectstest.PassUntilTaxed(t, g, me.ID)
	if !g.Stack.Contains(victim) {
		t.Fatal("the spell resolved while the {4} was unanswered")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Stack.Contains(victim) {
		t.Error("declining the tax must counter the spell")
	}
}

func TestDftSZSpectralInterferenceRefusesAnInstant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Instant", "Instant", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	them.Hand.PushTop(game.Card{InstanceID: id, Name: "Spectral Interference", TypeLine: "Instant",
		OracleID: dftSZSpectralOracle, Owner: them.ID, Controller: them.ID})
	if err := g.CastSpell(them.ID, id, game.CastSpellParams{Targets: effectstest.DftABCardRef(victim)}); err == nil {
		t.Error("Spectral Interference targeted an instant spell")
	}
}

// --- Skybox Ferry -----------------------------------------------------------------

func TestDftSZSkyboxFerryCrewsTwoAndFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ferry := effectstest.DftABPush(g, me.ID, "Skybox Ferry", effectstest.DftABVehicle, dftSZSkyboxFerryOracle, 4, 4)
	one := effectstest.B12Creature(g, me.ID, "One", "Creature — Rat", 1, 1)
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ferry, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{one}}); err == nil {
		t.Fatal("power 1 crewed a Crew 2 Vehicle")
	}
	effectstest.DftABCrew(t, g, me, ferry, 0, two)
	if !effectstest.HasEffectiveKeyword(t, g, ferry, "flying") {
		t.Error("the crewed Ferry has no flying")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, ferry); p != 4 || tough != 4 {
		t.Errorf("the crewed Ferry is %d/%d, want 4/4", p, tough)
	}
}

func TestDftSZSkyboxFerryCyclesForTwo(t *testing.T) {
	dftSZCycles(t, "Skybox Ferry", effectstest.DftABVehicle, dftSZSkyboxFerryOracle, "{C}{C}", 1)
}

// --- Spotcycle Scouter ------------------------------------------------------------

func TestDftSZSpotcycleScouterScriesTwoAndCrewsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "Top", "Second", "Third")
	scouter := effectstest.DftABCast(t, g, me, game.Card{Name: "Spotcycle Scouter", TypeLine: effectstest.DftABVehicle,
		OracleID: dftSZSpotcycleOracle, Power: 3, Toughness: 2})
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt after the Scouter entered")
	}
	if len(c.ScryCards) != 2 {
		t.Fatalf("looked at %d cards, want 2", len(c.ScryCards))
	}
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LibraryTopNames(me, 1); got[0] != "Third" {
		t.Errorf("library top is %q, want Third after bottoming two", got[0])
	}

	// Crew is not a {T} ability of the Vehicle: the Scouter crews the
	// turn it enters.
	pilot := effectstest.B12Creature(g, me.ID, "Pilot", "Creature — Human Pilot", 1, 1)
	effectstest.DftABCrew(t, g, me, scouter, 0, pilot)
}
