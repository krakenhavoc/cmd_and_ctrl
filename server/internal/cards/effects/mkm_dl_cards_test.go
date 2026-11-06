package effects

import (
	"testing"

	"github.com/google/uuid"

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
		if ac.ManaCost != MorphCastCost || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s: cast %q, kind %q, face-up %q; want %q, disguised, %q",
				tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCost, MorphCastCost, tc.faceUp)
		}
	}
}

// --- Defenestrated Phantom -----------------------------------------------------

func TestMkmDLDefenestratedPhantomFliesOnlyFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmMPCastFaceDown(t, g, "Defenestrated Phantom", "Creature — Spirit", mkmDLPhantomOracle, 4, 3)
	if hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("the face-down Phantom flies (CR 708.2a)")
	}
	mkmMPTurnFaceUp(t, g, me, id, "W", "C", "C", "C", "C")
	assertKeywords(t, g, id, "flying")
	if p, tough := mkmMPPT(t, g, id); p != 4 || tough != 3 {
		t.Errorf("face up the Phantom is %d/%d, want 4/3", p, tough)
	}
}

// --- Dog Walker ----------------------------------------------------------------

func TestMkmDLDogWalkerMakesTwoTappedDogsWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmMPCastFaceDown(t, g, "Dog Walker", "Creature — Human Citizen", mkmDLDogWalkerOracle, 3, 1)
	if got := tmtB1Count(g, me.ID, "Dog"); got != 0 {
		t.Fatalf("%d Dogs from entering face down, want 0", got)
	}
	mkmMPTurnFaceUp(t, g, me, id, "R", "W")
	passPriorityAroundTable(t, g)

	dogs := tmtB1Named(g, "Dog")
	if len(dogs) != 2 {
		t.Fatalf("%d Dogs, want 2", len(dogs))
	}
	for _, dog := range dogs {
		c := tmtB1Card(t, g, dog)
		if !c.Tapped {
			t.Error("a Dog entered untapped")
		}
		if c.Controller != me.ID || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W" {
			t.Errorf("a Dog is %d/%d %v under %s, want my 1/1 white", c.Power, c.Toughness, c.Colors, c.Controller)
		}
	}
	assertKeywords(t, g, id, "vigilance")
}

func TestMkmDLDogWalkerCastFaceUpMakesNoDogs(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Dog Walker", "Creature — Human Citizen", mkmDLDogWalkerOracle, 3, 1)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Dog"); got != 0 {
		t.Errorf("%d Dogs from a face-up cast, want 0", got)
	}
}

// --- Exit Specialist -----------------------------------------------------------

func TestMkmDLExitSpecialistCantBeBlockedByPowerThreeOrGreater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exit := b12Push(g, me.ID, "Exit Specialist", "Creature — Human Detective", mkmDLExitOracle, 2, 1)
	big := b12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	declareAttack(t, g, opp.ID, exit)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, exit); err == nil {
		t.Error("a 3/3 blocked the Exit Specialist")
	}
	if err := g.DeclareBlocker(small, exit); err != nil {
		t.Errorf("a 2/2 could not block the Exit Specialist: %v", err)
	}
}

func TestMkmDLExitSpecialistBouncesAnotherCreatureWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := mkmMPCastFaceDown(t, g, "Exit Specialist", "Creature — Human Detective", mkmDLExitOracle, 2, 1)

	mkmMPTurnFaceUp(t, g, me, id, "U", "C")
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("turning the Specialist face up asked for no target")
	}
	if hasID(p.PickTargetCards, id) {
		t.Error("\"another\": the Specialist was offered as its own target")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the target did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the Specialist left the battlefield")
	}
}

// --- Granite Witness -----------------------------------------------------------

func TestMkmDLGraniteWitnessTapsATargetWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := mkmMPCastFaceDown(t, g, "Granite Witness", "Artifact Creature — Gargoyle Detective", mkmDLGraniteOracle, 3, 2)
	if hasKeywordOnBattlefield(t, g, id, "flying") {
		t.Error("the face-down Witness flies (CR 708.2a)")
	}

	mkmMPTurnFaceUp(t, g, me, id, "W", "U")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, theirs)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	if !b16Tapped(t, g, theirs) {
		t.Error("choosing \"tap it\" left the target untapped")
	}
	assertKeywords(t, g, id, "flying", "vigilance")
}

func TestMkmDLGraniteWitnessUntapsATarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	id := mkmMPCastFaceDown(t, g, "Granite Witness", "Artifact Creature — Gargoyle Detective", mkmDLGraniteOracle, 3, 2)

	mkmMPTurnFaceUp(t, g, me, id, "W", "W")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, mine)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 1)
	if b16Tapped(t, g, mine) {
		t.Error("choosing \"untap it\" left the target tapped")
	}
}

// --- Greenbelt Radical ---------------------------------------------------------

func TestMkmDLGreenbeltRadicalGrowsAndTramplesYourTeamWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := mkmMPCastFaceDown(t, g, "Greenbelt Radical", "Creature — Centaur Citizen", mkmDLGreenbeltOracle, 4, 4)
	if counterCount(g, mine, game.CounterPlusOne) != 0 {
		t.Fatal("entering face down put a counter on a creature")
	}

	mkmMPTurnFaceUp(t, g, me, id, "G", "G", "C", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	for _, c := range []struct {
		name string
		id   uuid.UUID
	}{{"my bear", mine}, {"the Radical", id}} {
		if got := counterCount(g, c.id, game.CounterPlusOne); got != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", c.name, got)
		}
		if !hasEffectiveKeyword(t, g, c.id, "trample") {
			t.Errorf("%s has no trample", c.name)
		}
	}
	if p, tough := mkmMPPT(t, g, id); p != 5 || tough != 5 {
		t.Errorf("the Radical is %d/%d, want 5/5", p, tough)
	}
	if counterCount(g, theirs, game.CounterPlusOne) != 0 || hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature was grown or given trample")
	}
}

// --- Gadget Technician ---------------------------------------------------------

func TestMkmDLGadgetTechnicianMakesAThopterOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Gadget Technician", "Creature — Goblin Artificer", mkmDLGadgetOracle, 3, 2)
	passPriorityAroundTable(t, g)
	thopters := tmtB1Named(g, "Thopter")
	if len(thopters) != 1 {
		t.Fatalf("%d Thopters, want 1", len(thopters))
	}
	if c := tmtB1Card(t, g, thopters[0]); c.Controller != me.ID || c.Power != 1 || c.Toughness != 1 || !c.IsArtifact() {
		t.Errorf("the Thopter is %d/%d (artifact %v) under %s", c.Power, c.Toughness, c.IsArtifact(), c.Controller)
	}
	assertKeywords(t, g, thopters[0], "flying")
}

func TestMkmDLGadgetTechnicianMakesAThopterWhenTurnedFaceUpNotWhenEnteringFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmMPCastFaceDown(t, g, "Gadget Technician", "Creature — Goblin Artificer", mkmDLGadgetOracle, 3, 2)
	if got := tmtB1Count(g, me.ID, "Thopter"); got != 0 {
		t.Fatalf("%d Thopters from entering face down, want 0 (CR 708.3)", got)
	}
	mkmMPTurnFaceUp(t, g, me, id, "U", "R")
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Thopter"); got != 1 {
		t.Errorf("%d Thopters after turning it face up, want 1", got)
	}
}

// --- Fanatical Strength --------------------------------------------------------

func TestMkmDLFanaticalStrengthPumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Fanatical Strength", "Instant", mkmDLFanaticalOracle, []game.TargetRef{cardRef(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the target is %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target has no trample")
	}
}

// --- Long Goodbye --------------------------------------------------------------

func TestMkmDLLongGoodbyeDestroysASmallCreatureOrPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	three := b27Push(g, opp.ID, "Their Three", "Creature — Bear", "", "{2}{G}", 3, 3, "G")
	four := b27Push(g, opp.ID, "Their Four", "Creature — Elk", "", "{3}{G}", 4, 4, "G")
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{1}{U}{U}",
		Colors: []string{"U"}, Counters: map[string]int{game.CounterLoyalty: 3}, Owner: opp.ID, Controller: opp.ID,
	})

	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Long Goodbye", TypeLine: "Instant",
		OracleID: mkmDLGoodbyeOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: []game.TargetRef{cardRef(four)}}); err == nil {
		t.Fatal("a mana value 4 creature was accepted as the target")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: []game.TargetRef{cardRef(three)}}); err != nil {
		t.Fatalf("cast Long Goodbye: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(three) {
		t.Error("the mana value 3 creature survived")
	}
	if !g.Battlefield.Contains(four) {
		t.Error("the mana value 4 creature left the battlefield")
	}

	castCatalogSpell(t, g, "Long Goodbye", "Instant", mkmDLGoodbyeOracle, []game.TargetRef{cardRef(walker)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the mana value 3 planeswalker survived")
	}
}

func TestMkmDLLongGoodbyeCantBeCountered(t *testing.T) {
	spec, ok := Lookup(mkmDLGoodbyeOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Long Goodbye does not declare that it can't be countered")
	}
	if game.CatalogCantBeCountered == nil || !game.CatalogCantBeCountered(mkmDLGoodbyeOracle) {
		t.Error("the engine's can't-be-countered hook does not see Long Goodbye")
	}
}

// --- Hard-Hitting Question -----------------------------------------------------

func TestMkmDLHardHittingQuestionBitesWithTheCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	castCatalogSpell(t, g, "Hard-Hitting Question", "Sorcery", mkmDLQuestionOracle, cardRefs(mine, theirs))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 3/3 survived 3 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my creature took damage back: the bite is one-sided")
	}
}

func TestMkmDLHardHittingQuestionRefusesYourOwnCreatureAsTheVictim(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := b12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	b := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Hard-Hitting Question", TypeLine: "Sorcery",
		OracleID: mkmDLQuestionOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Targets: cardRefs(a, b)}); err == nil {
		t.Error("my own creature was accepted as the creature I don't control")
	}
}

// --- Kraul Whipcracker ---------------------------------------------------------

func TestMkmDLKraulWhipcrackerDestroysAnOpponentsToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	token := pushToken(g, opp.ID, TokenCard("2/2 white and blue Detective"))
	mine := pushToken(g, me.ID, TokenCard("2/2 white and blue Detective"))
	card := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	id := castHoldCreature(t, g, "Kraul Whipcracker", "Creature — Insect Assassin", mkmDLWhipcrackerOracle, 3, 2)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if hasID(p.PickTargetCards, mine) || hasID(p.PickTargetCards, card) {
		t.Error("my token or a nontoken creature was offered")
	}
	pickCard(t, g, me.ID, token)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the opponent's token survived")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(card) {
		t.Error("something else was destroyed")
	}
	assertKeywords(t, g, id, "reach")
}

func TestMkmDLKraulWhipcrackerWithNoOpposingTokenAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushToken(g, me.ID, TokenCard("2/2 white and blue Detective"))
	id := castHoldCreature(t, g, "Kraul Whipcracker", "Creature — Insect Assassin", mkmDLWhipcrackerOracle, 3, 2)
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, id) != nil {
		t.Error("the trigger waited on a target with no opposing token (CR 603.3d)")
	}
}

// --- Haazda Vigilante ----------------------------------------------------------

func TestMkmDLHaazdaVigilanteCountersASmallCreatureOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)

	id := castHoldCreature(t, g, "Haazda Vigilante", "Creature — Giant Soldier", mkmDLVigilanteOracle, 4, 4)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if hasID(p.PickTargetCards, big) || hasID(p.PickTargetCards, id) {
		t.Error("a creature with power 3 or greater was offered")
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, small, game.CounterPlusOne); got != 1 {
		t.Errorf("the target has %d counters after the enters trigger, want 1", got)
	}
}

func TestMkmDLHaazdaVigilanteCountersASmallCreatureWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, "Haazda Vigilante", "Creature — Giant Soldier", mkmDLVigilanteOracle, 4, 4)
	big := b12Creature(g, me.ID, "My Elk", "Creature — Elk", 3, 3)
	other := b12Creature(g, me.ID, "My Mouse", "Creature — Mouse", 1, 1)

	declareAttack(t, g, opp.ID, id)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the attack trigger asked for no target")
	}
	if hasID(p.PickTargetCards, big) {
		t.Error("a 3/3 was offered")
	}
	pickCard(t, g, me.ID, other)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, other, game.CounterPlusOne); got != 1 {
		t.Errorf("the attack trigger's target has %d counters, want 1", got)
	}
}

// --- Inside Source -------------------------------------------------------------

func TestMkmDLInsideSourceMakesADetectiveOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castHoldCreature(t, g, "Inside Source", "Creature — Human Citizen", mkmDLSourceOracle, 1, 1)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Detective"); got != 1 {
		t.Errorf("%d Detectives, want 1", got)
	}
}

func TestMkmDLInsideSourcePumpsATargetDetective(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	source := b12Push(g, me.ID, "Inside Source", "Creature — Human Citizen", mkmDLSourceOracle, 1, 1)
	detective := pushToken(g, me.ID, TokenCard("2/2 white and blue Detective"))
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C", "C", "C")...) })

	if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{cardRef(bear)}}); err == nil {
		t.Fatal("a creature that is not a Detective was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, source, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{cardRef(detective)}}); err != nil {
		t.Fatalf("activate {3}, {T}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, detective); p != 4 || tough != 2 {
		t.Errorf("the Detective is %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, detective, "vigilance") {
		t.Error("the Detective has no vigilance")
	}
	if !tmtB1Card(t, g, source).Tapped {
		t.Error("the Source is untapped after its {T} ability")
	}
}
