package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestMkmMPEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.MkmMPGlassOracle:      "Magnifying Glass",
		effectstest.MkmMPBindingOracle:    "Makeshift Binding",
		effectstest.MkmMPPhantomOracle:    "Marketwatch Phantom",
		effectstest.MkmMPGuardianOracle:   "Neighborhood Guardian",
		effectstest.MkmMPWatchOracle:      "Not on My Watch",
		effectstest.MkmMPInspectorOracle:  "Novice Inspector",
		effectstest.MkmMPOnTheJobOracle:   "On the Job",
		effectstest.MkmMPNightwatchOracle: "Museum Nightwatch",
		effectstest.MkmMPMoroiiOracle:     "Nightdrinker Moroii",
		effectstest.MkmMPOffenderOracle:   "Offender at Large",
		effectstest.MkmMPGardenerOracle:   "Nervous Gardener",
		effectstest.MkmMPPerformerOracle:  "Pyrotechnic Performer",
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

// TestMkmMPDisguiseDeclarationsAreWired pins each disguise card's
// keyword: the {3} face-down cast, the disguise kind (ward {2}) and the
// card's own face-up cost.
func TestMkmMPDisguiseDeclarationsAreWired(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Museum Nightwatch", effectstest.MkmMPNightwatchOracle, "{1}{W}"},
		{"Nightdrinker Moroii", effectstest.MkmMPMoroiiOracle, "{B}{B}"},
		{"Offender at Large", effectstest.MkmMPOffenderOracle, "{4}{R}"},
		{"Nervous Gardener", effectstest.MkmMPGardenerOracle, "{G}"},
		{"Pyrotechnic Performer", effectstest.MkmMPPerformerOracle, "{R}"},
	} {
		ac := game.AlternativeCostByKey(tc.oracle, "disguise")
		if ac == nil || ac.FaceDown == nil {
			t.Errorf("%s offers no disguise cast", tc.name)
			continue
		}
		if ac.ManaCost != effects.MorphCastCost || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s: cast %q, kind %q, face-up %q; want %q, disguised, %q",
				tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCost, effects.MorphCastCost, tc.faceUp)
		}
	}
}

// --- Magnifying Glass ----------------------------------------------------------

func TestMkmMPMagnifyingGlassTapsForColorlessAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	glass := effectstest.B12Push(g, me.ID, "Magnifying Glass", "Artifact", effectstest.MkmMPGlassOracle, 0, 0)

	if err := g.ActivateManaAbility(me.ID, glass, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := me.ManaPool.CountColor("C"); got != 1 {
		t.Errorf("%d colorless in the pool, want 1", got)
	}
	if !effectstest.TmtB1Card(t, g, glass).Tapped {
		t.Error("the Glass is untapped after its mana ability")
	}
}

func TestMkmMPMagnifyingGlassInvestigatesForFourAndATap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	glass := effectstest.B12Push(g, me.ID, "Magnifying Glass", "Artifact", effectstest.MkmMPGlassOracle, 0, 0)
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("C", "C", "C", "C")...) })

	if err := g.ActivateCatalogAbility(me.ID, glass, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate {4}, {T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
	if !effectstest.TmtB1Card(t, g, glass).Tapped {
		t.Error("the Glass is untapped after its {T} ability")
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left in the pool, want the {4} spent", got)
	}
}

// --- Makeshift Binding ---------------------------------------------------------

func TestMkmMPMakeshiftBindingExilesUntilItLeavesAndGainsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life

	binding := effectstest.CastHoldCreature(t, g, "Makeshift Binding", "Enchantment", effectstest.MkmMPBindingOracle, 0, 0)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) || effectstest.OnBattlefieldNamed(g, "Their Bear") != 0 {
		t.Fatal("the creature was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(binding) })
	g.RunStateChecksForTest()
	if effectstest.TmtB1Count(g, opp.ID, "Their Bear") != 1 {
		t.Error("the creature did not return when the Binding left")
	}
}

func TestMkmMPMakeshiftBindingOffersOnlyAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Makeshift Binding", "Enchantment", effectstest.MkmMPBindingOracle, 0, 0)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
			t.Error("my own creature was accepted as the target")
		}
	}
}

// --- Marketwatch Phantom -------------------------------------------------------

func TestMkmMPMarketwatchPhantomFliesWhenASmallCreatureEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	phantom := effectstest.B12Push(g, me.ID, "Marketwatch Phantom", "Creature — Spirit Detective", effectstest.MkmMPPhantomOracle, 2, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, phantom, "flying") {
		t.Fatal("the Phantom flies before anything entered")
	}

	effectstest.CastHoldCreature(t, g, "Neighborhood Guardian", "Creature — Unicorn", effectstest.MkmMPGuardianOracle, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, phantom, "flying") {
		t.Error("a 2/2 entered and the Phantom has no flying")
	}
}

func TestMkmMPMarketwatchPhantomIgnoresAThreePowerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	phantom := effectstest.B12Push(g, me.ID, "Marketwatch Phantom", "Creature — Spirit Detective", effectstest.MkmMPPhantomOracle, 2, 2)
	effectstest.CastHoldCreature(t, g, "Pyrotechnic Performer", effectstest.MkmMPPerformerTypeLine, effectstest.MkmMPPerformerOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, phantom, "flying") {
		t.Error("a 3/2 entering gave the Phantom flying")
	}
}

func TestMkmMPMarketwatchPhantomCountsAFaceDownTwoTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	phantom := effectstest.B12Push(g, me.ID, "Marketwatch Phantom", "Creature — Spirit Detective", effectstest.MkmMPPhantomOracle, 2, 2)
	// The Performer is 3/2 face up and a 2/2 face down (CR 708.2a).
	effectstest.MkmMPCastFaceDown(t, g, "Pyrotechnic Performer", effectstest.MkmMPPerformerTypeLine, effectstest.MkmMPPerformerOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, phantom, "flying") {
		t.Error("a face-down 2/2 entered and the Phantom has no flying")
	}
}

// --- Neighborhood Guardian -----------------------------------------------------

func TestMkmMPNeighborhoodGuardianPumpsTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	guardian := effectstest.B12Push(g, me.ID, "Neighborhood Guardian", "Creature — Unicorn", effectstest.MkmMPGuardianOracle, 2, 2)

	effectstest.CastHoldCreature(t, g, "Marketwatch Phantom", "Creature — Spirit Detective", effectstest.MkmMPPhantomOracle, 2, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, guardian)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, guardian); p != 3 || tough != 3 {
		t.Errorf("the Guardian is %d/%d, want 3/3", p, tough)
	}
}

func TestMkmMPNeighborhoodGuardianIgnoresAThreePowerCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Neighborhood Guardian", "Creature — Unicorn", effectstest.MkmMPGuardianOracle, 2, 2)
	effectstest.CastHoldCreature(t, g, "Pyrotechnic Performer", effectstest.MkmMPPerformerTypeLine, effectstest.MkmMPPerformerOracle, 3, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a 3/2 entering triggered the Guardian")
	}
}

// --- Not on My Watch -----------------------------------------------------------

func TestMkmMPNotOnMyWatchExilesAnAttackingCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.B12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)

	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Not on My Watch", TypeLine: "Instant",
		OracleID: effectstest.MkmMPWatchOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}},
	}); err == nil {
		t.Fatal("a creature that is not attacking was accepted as the target")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}},
	}); err != nil {
		t.Fatalf("cast Not on My Watch: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(attacker) || !g.Exile.Contains(attacker) {
		t.Error("the attacking creature was not exiled")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the other creature left the battlefield")
	}
}

// --- Novice Inspector ----------------------------------------------------------

func TestMkmMPNoviceInspectorInvestigatesOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Novice Inspector", "Creature — Human Detective", effectstest.MkmMPInspectorOracle, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- On the Job ----------------------------------------------------------------

func TestMkmMPOnTheJobPumpsYourCreaturesAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "On the Job", "Instant", effectstest.MkmMPOnTheJobOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.MkmMPPT(t, g, a); p != 4 || tough != 3 {
		t.Errorf("my 2/2 is %d/%d, want 4/3", p, tough)
	}
	if p, tough := effectstest.MkmMPPT(t, g, b); p != 5 || tough != 4 {
		t.Errorf("my 3/3 is %d/%d, want 5/4", p, tough)
	}
	if p, tough := effectstest.MkmMPPT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("the opponent's 2/2 is %d/%d, want unchanged", p, tough)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Museum Nightwatch ---------------------------------------------------------

func TestMkmMPMuseumNightwatchMakesADetectiveWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	watch := effectstest.B12Push(g, me.ID, "Museum Nightwatch", "Creature — Centaur Soldier", effectstest.MkmMPNightwatchOracle, 3, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(watch) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives, want 1", got)
	}
}

func TestMkmMPMuseumNightwatchDyingFaceDownMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	watch := effectstest.MkmMPCastFaceDown(t, g, "Museum Nightwatch", "Creature — Centaur Soldier", effectstest.MkmMPNightwatchOracle, 3, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(watch) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Detective"); got != 0 {
		t.Errorf("%d Detectives from a face-down Nightwatch, want 0 (CR 708.2a)", got)
	}
}

func TestMkmMPMuseumNightwatchTurnedFaceUpIsAThreeTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	watch := effectstest.MkmMPCastFaceDown(t, g, "Museum Nightwatch", "Creature — Centaur Soldier", effectstest.MkmMPNightwatchOracle, 3, 2)
	effectstest.MkmMPTurnFaceUp(t, g, me, watch, "W", "C")
	if p, tough := effectstest.MkmMPPT(t, g, watch); p != 3 || tough != 2 {
		t.Errorf("face up the Nightwatch is %d/%d, want 3/2", p, tough)
	}
}

// --- Nightdrinker Moroii -------------------------------------------------------

func TestMkmMPNightdrinkerMoroiiFliesAndCostsThreeLifeOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := effectstest.CastHoldCreature(t, g, "Nightdrinker Moroii", "Creature — Vampire", effectstest.MkmMPMoroiiOracle, 4, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, me.Life)
	}
	effectstest.AssertKeywords(t, g, id, "flying")
}

func TestMkmMPNightdrinkerMoroiiFaceDownNeverLosesTheLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	id := effectstest.MkmMPCastFaceDown(t, g, "Nightdrinker Moroii", "Creature — Vampire", effectstest.MkmMPMoroiiOracle, 4, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("the face-down Moroii kept flying (CR 708.2a)")
	}
	effectstest.MkmMPTurnFaceUp(t, g, me, id, "B", "B")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d -> %d; neither the face-down entry nor turning it face up loses life (CR 708.3, CR 708.8)", life, me.Life)
	}
	effectstest.AssertKeywords(t, g, id, "flying")
	if p, tough := effectstest.MkmMPPT(t, g, id); p != 4 || tough != 2 {
		t.Errorf("face up the Moroii is %d/%d, want 4/2", p, tough)
	}
}

// --- Offender at Large ---------------------------------------------------------

func TestMkmMPOffenderAtLargePumpsOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Offender at Large", "Creature — Giant Rogue", effectstest.MkmMPOffenderOracle, 5, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
}

func TestMkmMPOffenderAtLargePumpsWhenTurnedFaceUpNotWhenEnteringFaceDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := effectstest.MkmMPCastFaceDown(t, g, "Offender at Large", "Creature — Giant Rogue", effectstest.MkmMPOffenderOracle, 5, 4)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Offender")
	}
	effectstest.MkmMPTurnFaceUp(t, g, me, id, "R", "C", "C", "C", "C")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("the target is %d/%d, want 4/2", p, tough)
	}
}

func TestMkmMPOffenderAtLargeMayTargetNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Offender at Large", "Creature — Giant Rogue", effectstest.MkmMPOffenderOracle, 5, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B26DeclinePickTarget(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("with no target chosen the bear is %d/%d, want 2/2", p, tough)
	}
}

// --- Nervous Gardener ----------------------------------------------------------

func TestMkmMPNervousGardenerFetchesALandWithABasicLandTypeWhenTurnedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dual := effectstest.StapleLibraryCard(me, "Breeding Pool", "Land — Forest Island")
	effectstest.StapleLibraryCard(me, "Wastes", "Basic Land")
	id := effectstest.MkmMPCastFaceDown(t, g, "Nervous Gardener", "Creature — Dryad", effectstest.MkmMPGardenerOracle, 2, 2)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("entering face down asked for a search")
	}

	effectstest.MkmMPTurnFaceUp(t, g, me, id, "G")
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("turning the Gardener face up asked for no search")
	}
	if effectstest.SearchOptionNamed(g, c, "Wastes") != uuid.Nil {
		t.Error("Wastes, which has no basic land type, was offered")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, dual)
	if !me.Hand.Contains(dual) {
		t.Error("the land did not reach the hand")
	}
}

// --- Pyrotechnic Performer -----------------------------------------------------

func TestMkmMPPyrotechnicPerformerBurnsEachOpponentWhenItTurnsUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := effectstest.MkmMPOpponents(g)
	id := effectstest.MkmMPCastFaceDown(t, g, "Pyrotechnic Performer", effectstest.MkmMPPerformerTypeLine, effectstest.MkmMPPerformerOracle, 3, 2)
	before := effectstest.MkmMPLives(opps)
	myLife := me.Life

	effectstest.MkmMPTurnFaceUp(t, g, me, id, "R")
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range opps {
		if p.Life != before[i]-3 {
			t.Errorf("opponent %d: life %d -> %d, want -3", i, before[i], p.Life)
		}
	}
	if me.Life != myLife {
		t.Errorf("my life changed %d -> %d", myLife, me.Life)
	}
}

func TestMkmMPPyrotechnicPerformerUsesTheOtherCreaturesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := effectstest.MkmMPOpponents(g)
	effectstest.B12Push(g, me.ID, "Pyrotechnic Performer", effectstest.MkmMPPerformerTypeLine, effectstest.MkmMPPerformerOracle, 3, 2)
	moroii := effectstest.MkmMPCastFaceDown(t, g, "Nightdrinker Moroii", "Creature — Vampire", effectstest.MkmMPMoroiiOracle, 4, 2)
	before := effectstest.MkmMPLives(opps)

	effectstest.MkmMPTurnFaceUp(t, g, me, moroii, "B", "B")
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range opps {
		if p.Life != before[i]-4 {
			t.Errorf("opponent %d: life %d -> %d, want -4 (the Moroii's power)", i, before[i], p.Life)
		}
	}
}
