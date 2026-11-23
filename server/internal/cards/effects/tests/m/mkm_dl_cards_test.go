package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_dl_cards_test.go — Murders at Karlov Manor cards D–L (#4469). Each
// test plays the real catalog entry and asserts what its text says. The
// face-down helpers are the M–P batch's (mkmMPCastFaceDown,
// mkmMPTurnFaceUp, mkmMPPT).

const (
	mkmDLPhantomOracle     = "e73addfa-d18a-46f7-882e-7bc2f292230b"
	mkmDLDogWalkerOracle   = "10d24900-cd40-4c98-bb1e-b9e66113f703"
	mkmDLExitOracle        = "79d2efb2-145d-4377-a4e2-832ff2cb5677"
	mkmDLGraniteOracle     = "37f2ec12-d703-40cd-a155-053ae2e4ddcc"
	mkmDLGreenbeltOracle   = "d0a4efc3-fa68-48f2-a6d0-2fff93e05c77"
	mkmDLGadgetOracle      = "d7f777ae-6122-4af9-b962-7574bc8e2691"
	mkmDLFanaticalOracle   = "b92e97f5-a307-471f-be06-afa4fcdab757"
	mkmDLGoodbyeOracle     = "9d63ba8a-a5cd-4fdb-bba3-95612be0077c"
	mkmDLQuestionOracle    = "05fdefb4-2be7-4d9a-9ab9-14bbbe72aef7"
	mkmDLWhipcrackerOracle = "b563ae31-338e-4bcc-a05f-d7ed7c3fe46e"
	mkmDLVigilanteOracle   = "4b707ef4-2c17-41e7-8dac-583791a98055"
	mkmDLSourceOracle      = "12bced67-69bc-4fc7-a585-577060afdd50"
)

func TestMkmDLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmDLPhantomOracle:     "Defenestrated Phantom",
		mkmDLDogWalkerOracle:   "Dog Walker",
		mkmDLExitOracle:        "Exit Specialist",
		mkmDLGraniteOracle:     "Granite Witness",
		mkmDLGreenbeltOracle:   "Greenbelt Radical",
		mkmDLGadgetOracle:      "Gadget Technician",
		mkmDLFanaticalOracle:   "Fanatical Strength",
		mkmDLGoodbyeOracle:     "Long Goodbye",
		mkmDLQuestionOracle:    "Hard-Hitting Question",
		mkmDLWhipcrackerOracle: "Kraul Whipcracker",
		mkmDLVigilanteOracle:   "Haazda Vigilante",
		mkmDLSourceOracle:      "Inside Source",
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

// TestMkmDLDisguiseDeclarationsAreWired pins each disguise card's keyword:
// the {3} face-down cast, the disguise kind (ward {2}) and the card's own
// face-up cost.
func TestMkmDLDisguiseDeclarationsAreWired(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Defenestrated Phantom", mkmDLPhantomOracle, "{4}{W}"},
		{"Dog Walker", mkmDLDogWalkerOracle, "{R/W}{R/W}"},
		{"Exit Specialist", mkmDLExitOracle, "{1}{U}"},
		{"Granite Witness", mkmDLGraniteOracle, "{W/U}{W/U}"},
		{"Greenbelt Radical", mkmDLGreenbeltOracle, "{5}{G}{G}"},
		{"Gadget Technician", mkmDLGadgetOracle, "{U/R}{U/R}"},
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

// --- Defenestrated Phantom -----------------------------------------------------

func TestMkmDLDefenestratedPhantomFliesOnlyFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.MkmMPCastFaceDown(t, g, "Defenestrated Phantom", "Creature — Spirit", mkmDLPhantomOracle, 4, 3)
	if effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("the face-down Phantom flies (CR 708.2a)")
	}
	effectstest.MkmMPTurnFaceUp(t, g, me, id, "W", "C", "C", "C", "C")
	effectstest.AssertKeywords(t, g, id, "flying")
	if p, tough := effectstest.MkmMPPT(t, g, id); p != 4 || tough != 3 {
		t.Errorf("face up the Phantom is %d/%d, want 4/3", p, tough)
	}
}

// --- Dog Walker ----------------------------------------------------------------

func TestMkmDLDogWalkerMakesTwoTappedDogsWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.MkmMPCastFaceDown(t, g, "Dog Walker", "Creature — Human Citizen", mkmDLDogWalkerOracle, 3, 1)
	if got := effectstest.TmtB1Count(g, me.ID, "Dog"); got != 0 {
		t.Fatalf("%d Dogs from entering face down, want 0", got)
	}
	effectstest.MkmMPTurnFaceUp(t, g, me, id, "R", "W")
	effectstest.PassPriorityAroundTable(t, g)

	dogs := effectstest.TmtB1Named(g, "Dog")
	if len(dogs) != 2 {
		t.Fatalf("%d Dogs, want 2", len(dogs))
	}
	for _, dog := range dogs {
		c := effectstest.TmtB1Card(t, g, dog)
		if !c.Tapped {
			t.Error("a Dog entered untapped")
		}
		if c.Controller != me.ID || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W" {
			t.Errorf("a Dog is %d/%d %v under %s, want my 1/1 white", c.Power, c.Toughness, c.Colors, c.Controller)
		}
	}
	effectstest.AssertKeywords(t, g, id, "vigilance")
}

func TestMkmDLDogWalkerCastFaceUpMakesNoDogs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Dog Walker", "Creature — Human Citizen", mkmDLDogWalkerOracle, 3, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Dog"); got != 0 {
		t.Errorf("%d Dogs from a face-up cast, want 0", got)
	}
}

// --- Exit Specialist -----------------------------------------------------------

func TestMkmDLExitSpecialistCantBeBlockedByPowerThreeOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exit := effectstest.B12Push(g, me.ID, "Exit Specialist", "Creature — Human Detective", mkmDLExitOracle, 2, 1)
	big := effectstest.B12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	effectstest.DeclareAttack(t, g, opp.ID, exit)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, exit); err == nil {
		t.Error("a 3/3 blocked the Exit Specialist")
	}
	if err := g.DeclareBlocker(small, exit); err != nil {
		t.Errorf("a 2/2 could not block the Exit Specialist: %v", err)
	}
}

func TestMkmDLExitSpecialistBouncesAnotherCreatureWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.MkmMPCastFaceDown(t, g, "Exit Specialist", "Creature — Human Detective", mkmDLExitOracle, 2, 1)

	effectstest.MkmMPTurnFaceUp(t, g, me, id, "U", "C")
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("turning the Specialist face up asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, id) {
		t.Error("\"another\": the Specialist was offered as its own target")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Specialist left the battlefield")
	}
}

// --- Granite Witness -----------------------------------------------------------

func TestMkmDLGraniteWitnessTapsATargetWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.MkmMPCastFaceDown(t, g, "Granite Witness", "Artifact Creature — Gargoyle Detective", mkmDLGraniteOracle, 3, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("the face-down Witness flies (CR 708.2a)")
	}

	effectstest.MkmMPTurnFaceUp(t, g, me, id, "W", "U")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	if !effectstest.B16Tapped(t, g, theirs) {
		t.Error("choosing \"tap it\" left the target untapped")
	}
	effectstest.AssertKeywords(t, g, id, "flying", "vigilance")
}

func TestMkmDLGraniteWitnessUntapsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	id := effectstest.MkmMPCastFaceDown(t, g, "Granite Witness", "Artifact Creature — Gargoyle Detective", mkmDLGraniteOracle, 3, 2)

	effectstest.MkmMPTurnFaceUp(t, g, me, id, "W", "W")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, mine)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if effectstest.B16Tapped(t, g, mine) {
		t.Error("choosing \"untap it\" left the target tapped")
	}
}

// --- Greenbelt Radical ---------------------------------------------------------

func TestMkmDLGreenbeltRadicalGrowsAndTramplesYourTeamWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.MkmMPCastFaceDown(t, g, "Greenbelt Radical", "Creature — Centaur Citizen", mkmDLGreenbeltOracle, 4, 4)
	if effectstest.CounterCount(g, mine, game.CounterPlusOne) != 0 {
		t.Fatal("entering face down put a counter on a creature")
	}

	effectstest.MkmMPTurnFaceUp(t, g, me, id, "G", "G", "C", "C", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range []struct {
		name string
		id   uuid.UUID
	}{{"my bear", mine}, {"the Radical", id}} {
		if got := effectstest.CounterCount(g, c.id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", c.name, got)
		}
		if !effectstest.HasEffectiveKeyword(t, g, c.id, "trample") {
			t.Errorf("%s has no trample", c.name)
		}
	}
	if p, tough := effectstest.MkmMPPT(t, g, id); p != 5 || tough != 5 {
		t.Errorf("the Radical is %d/%d, want 5/5", p, tough)
	}
	if effectstest.CounterCount(g, theirs, game.CounterPlusOne) != 0 || effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature was grown or given trample")
	}
}

// --- Gadget Technician ---------------------------------------------------------

func TestMkmDLGadgetTechnicianMakesAThopterOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Gadget Technician", "Creature — Goblin Artificer", mkmDLGadgetOracle, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	thopters := effectstest.TmtB1Named(g, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("%d Thopters, want 1", len(thopters))
	}
	if c := effectstest.TmtB1Card(t, g, thopters[0]); c.Controller != me.ID || c.Power != 1 || c.Toughness != 1 || !c.IsArtifact() {
		t.Errorf("the Thopter is %d/%d (artifact %v) under %s", c.Power, c.Toughness, c.IsArtifact(), c.Controller)
	}
	effectstest.AssertKeywords(t, g, thopters[0], "flying")
}

func TestMkmDLGadgetTechnicianMakesAThopterWhenTurnedFaceUpNotWhenEnteringFaceDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.MkmMPCastFaceDown(t, g, "Gadget Technician", "Creature — Goblin Artificer", mkmDLGadgetOracle, 3, 2)
	if got := effectstest.TmtB1Count(g, me.ID, "Thopter"); got != 0 {
		t.Fatalf("%d Thopters from entering face down, want 0 (CR 708.3)", got)
	}
	effectstest.MkmMPTurnFaceUp(t, g, me, id, "U", "R")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Thopter"); got != 1 {
		t.Errorf("%d Thopters after turning it face up, want 1", got)
	}
}

// --- Fanatical Strength --------------------------------------------------------

func TestMkmDLFanaticalStrengthPumpsAndGrantsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Fanatical Strength", "Instant", mkmDLFanaticalOracle, []game.TargetRef{effectstest.CardRef(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target has no trample")
	}
}

// --- Long Goodbye --------------------------------------------------------------

func TestMkmDLLongGoodbyeDestroysASmallCreatureOrPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	three := effectstest.B27Push(g, opp.ID, "Their Three", "Creature — Bear", "", "{2}{G}", 3, 3, "G")
	four := effectstest.B27Push(g, opp.ID, "Their Four", "Creature — Elk", "", "{3}{G}", 4, 4, "G")
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{1}{U}{U}",
		Colors: []string{"U"}, Counters: map[string]int{game.CounterLoyalty: 3}, Owner: opp.ID, Controller: opp.ID,
	})

	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Long Goodbye", TypeLine: "Instant",
		OracleID: mkmDLGoodbyeOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: []game.TargetRef{effectstest.CardRef(four)}}); err == nil {
		t.Fatal("a mana value 4 creature was accepted as the target")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: []game.TargetRef{effectstest.CardRef(three)}}); err != nil {
		t.Fatalf("cast Long Goodbye: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) {
		t.Error("the mana value 3 creature survived")
	}
	if !g.Battlefield.Contains(four) {
		t.Error("the mana value 4 creature left the battlefield")
	}

	effectstest.CastCatalogSpell(t, g, "Long Goodbye", "Instant", mkmDLGoodbyeOracle, []game.TargetRef{effectstest.CardRef(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the mana value 3 planeswalker survived")
	}
}

func TestMkmDLLongGoodbyeCantBeCountered(t *testing.T) {
	spec, ok := effects.Lookup(mkmDLGoodbyeOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Long Goodbye does not declare that it can't be countered")
	}
	if game.CatalogCantBeCountered == nil || !game.CatalogCantBeCountered(mkmDLGoodbyeOracle) {
		t.Error("the engine's can't-be-countered hook does not see Long Goodbye")
	}
}

// --- Hard-Hitting Question -----------------------------------------------------

func TestMkmDLHardHittingQuestionBitesWithTheCreaturesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Hard-Hitting Question", "Sorcery", mkmDLQuestionOracle, effectstest.CardRefs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 3/3 survived 3 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my creature took damage back: the bite is one-sided")
	}
}

func TestMkmDLHardHittingQuestionRefusesYourOwnCreatureAsTheVictim(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := effectstest.B12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	b := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Hard-Hitting Question", TypeLine: "Sorcery",
		OracleID: mkmDLQuestionOracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: effectstest.CardRefs(a, b)}); err == nil {
		t.Error("my own creature was accepted as the creature I don't control")
	}
}

// --- Kraul Whipcracker ---------------------------------------------------------

func TestMkmDLKraulWhipcrackerDestroysAnOpponentsToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	token := effectstest.PushToken(g, opp.ID, effects.TokenCard("2/2 white and blue Detective"))
	mine := effectstest.PushToken(g, me.ID, effects.TokenCard("2/2 white and blue Detective"))
	card := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := effectstest.CastHoldCreature(t, g, "Kraul Whipcracker", "Creature — Insect Assassin", mkmDLWhipcrackerOracle, 3, 2)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, mine) || effectstest.HasID(p.PickTargetCards, card) {
		t.Error("my token or a nontoken creature was offered")
	}
	effectstest.PickCard(t, g, me.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the opponent's token survived")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(card) {
		t.Error("something else was destroyed")
	}
	effectstest.AssertKeywords(t, g, id, "reach")
}

func TestMkmDLKraulWhipcrackerWithNoOpposingTokenAsksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushToken(g, me.ID, effects.TokenCard("2/2 white and blue Detective"))
	id := effectstest.CastHoldCreature(t, g, "Kraul Whipcracker", "Creature — Insect Assassin", mkmDLWhipcrackerOracle, 3, 2)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.TriggerOnStack(g, id) != nil {
		t.Error("the trigger waited on a target with no opposing token (CR 603.3d)")
	}
}

// --- Haazda Vigilante ----------------------------------------------------------

func TestMkmDLHaazdaVigilanteCountersASmallCreatureOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)

	id := effectstest.CastHoldCreature(t, g, "Haazda Vigilante", "Creature — Giant Soldier", mkmDLVigilanteOracle, 4, 4)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, big) || effectstest.HasID(p.PickTargetCards, id) {
		t.Error("a creature with power 3 or greater was offered")
	}
	effectstest.PickCard(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d counters after the enters trigger, want 1", got)
	}
}

func TestMkmDLHaazdaVigilanteCountersASmallCreatureWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B12Push(g, me.ID, "Haazda Vigilante", "Creature — Giant Soldier", mkmDLVigilanteOracle, 4, 4)
	big := effectstest.B12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	other := effectstest.B12Creature(g, me.ID, "My Mouse", "Creature — Mouse", 1, 1)

	effectstest.DeclareAttack(t, g, opp.ID, id)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, big) {
		t.Error("a 3/3 was offered")
	}
	effectstest.PickCard(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("the attack trigger's target has %d counters, want 1", got)
	}
}

// --- Inside Source -------------------------------------------------------------

func TestMkmDLInsideSourceMakesADetectiveOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastHoldCreature(t, g, "Inside Source", "Creature — Human Citizen", mkmDLSourceOracle, 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives, want 1", got)
	}
}

func TestMkmDLInsideSourcePumpsATargetDetective(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	source := effectstest.B12Push(g, me.ID, "Inside Source", "Creature — Human Citizen", mkmDLSourceOracle, 1, 1)
	detective := effectstest.PushToken(g, me.ID, effects.TokenCard("2/2 white and blue Detective"))
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.ManaPool.AddMana(effectstest.ManaTokens("C", "C", "C")...) })

	if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{effectstest.CardRef(bear)}}); err == nil {
		t.Fatal("a creature that is not a Detective was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{effectstest.CardRef(detective)}}); err != nil {
		t.Fatalf("activate {3}, {T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.MkmMPPT(t, g, detective); p != 4 || tough != 2 {
		t.Errorf("the Detective is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, detective, "vigilance") {
		t.Error("the Detective has no vigilance")
	}
	if !effectstest.TmtB1Card(t, g, source).Tapped {
		t.Error("the Source is untapped after its {T} ability")
	}
}
