package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_b1_cards_test.go — Murders at Karlov Manor batch 1 (#4469). Each
// test plays the real catalog entry and asserts what its text says.

const (
	mkmB1ArrivalOracle     = "b25d5b06-44cf-42bc-90dd-3b238f702361"
	mkmB1CrackerOracle     = "2e18ee5b-cda3-428b-be91-939c410256ac"
	mkmB1BasherOracle      = "c23617c9-53f9-4063-8241-b504e9b80cbe"
	mkmB1StalkerOracle     = "e335485a-80fc-40d4-8be6-7eabf2bab571"
	mkmB1AssailantOracle   = "57eba48e-87bd-4b54-8635-7348020f2ba4"
	mkmB1AmbusherOracle    = "c11fef0f-972d-4756-9902-fd3e3af00707"
	mkmB1BenthicOracle     = "5539d78c-1acc-46dd-ba6d-cc0cdad6ff44"
	mkmB1CrookOracle       = "4477eeab-a51a-4dc2-9c97-83f2551d8b54"
	mkmB1CrimsonPulseOracl = "955504bc-050a-42a2-b2e4-ab213183b2be"
	mkmB1FalconOracle      = "33b31fb0-71ce-408a-a256-2a5a3b085cb6"
)

func TestMkmB1EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mkmB1ArrivalOracle:     "Auspicious Arrival",
		mkmB1CrackerOracle:     "Cold Case Cracker",
		mkmB1BasherOracle:      "Bolrac-Clan Basher",
		mkmB1StalkerOracle:     "Basilica Stalker",
		mkmB1AssailantOracle:   "Alley Assailant",
		mkmB1AmbusherOracle:    "Culvert Ambusher",
		mkmB1BenthicOracle:     "Benthic Criminologists",
		mkmB1CrookOracle:       "Cornered Crook",
		mkmB1CrimsonPulseOracl: "Case of the Crimson Pulse",
		mkmB1FalconOracle:      "Case of the Filched Falcon",
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

// TestMkmB1DisguiseDeclarationsAreWired pins each disguise card's
// keyword: the {3} face-down cast, the disguise kind (ward {2}) and the
// card's own face-up cost.
func TestMkmB1DisguiseDeclarationsAreWired(t *testing.T) {
	for _, tc := range []struct{ name, oracle, faceUp string }{
		{"Bolrac-Clan Basher", mkmB1BasherOracle, "{3}{R}{R}"},
		{"Basilica Stalker", mkmB1StalkerOracle, "{4}{B}"},
		{"Alley Assailant", mkmB1AssailantOracle, "{4}{B}{B}"},
		{"Culvert Ambusher", mkmB1AmbusherOracle, "{4}{G}"},
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

// mkmB1PickPlayer answers the open pick_target prompt with a player.
func mkmB1PickPlayer(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if err := effectstest.TmcB7TryPickPlayer(t, g, chooser, target); err != nil {
		t.Fatalf("pick player: %v", err)
	}
}

// --- Auspicious Arrival --------------------------------------------------------

func TestMkmB1AuspiciousArrivalPumpsThenInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Auspicious Arrival", "Instant", mkmB1ArrivalOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)

	if p, tough := effectstest.MkmB1PT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Cold Case Cracker ---------------------------------------------------------

func TestMkmB1ColdCaseCrackerFliesAndInvestigatesWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cracker := effectstest.B12Push(g, me.ID, "Cold Case Cracker", "Creature — Spirit Detective", mkmB1CrackerOracle, 3, 3)
	effectstest.AssertKeywords(t, g, cracker, "flying")

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cracker) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues after the Cracker died, want 1", got)
	}
}

// --- Bolrac-Clan Basher --------------------------------------------------------

func TestMkmB1BolracClanBasherKeywordsAndDisguise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.MkmB1CastFaceDown(t, g, "Bolrac-Clan Basher", "Creature — Cyclops Warrior", mkmB1BasherOracle, 3, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, id, "double strike") || effectstest.HasKeywordOnBattlefield(t, g, id, "trample") {
		t.Error("the face-down Basher kept a printed keyword (CR 708.2a)")
	}
	effectstest.MkmB1TurnFaceUp(t, g, me, id, "R", "R", "C", "C", "C")
	effectstest.AssertKeywords(t, g, id, "double strike", "trample")
	if p, tough := effectstest.MkmB1PT(t, g, id); p != 3 || tough != 2 {
		t.Errorf("face up the Basher is %d/%d, want 3/2", p, tough)
	}
}

// --- Basilica Stalker ----------------------------------------------------------

func TestMkmB1BasilicaStalkerGainsOneAndSurveilsOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stalker := effectstest.B12Push(g, me.ID, "Basilica Stalker", "Creature — Vampire Detective", mkmB1StalkerOracle, 3, 4)
	effectstest.AssertKeywords(t, g, stalker, "flying")
	life := me.Life

	effectstest.AttackWith(t, g, opp.ID, stalker)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Life; got != life+1 {
		t.Errorf("life %d -> %d, want +1", life, got)
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil {
		t.Fatal("no surveil 1 prompt")
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("surveil looks at %d cards, want 1", len(c.ScryCards))
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card did not go to the graveyard")
	}
}

// --- Alley Assailant -----------------------------------------------------------

func TestMkmB1AlleyAssailantEntersTappedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.CastHoldCreature(t, g, "Alley Assailant", "Creature — Vampire Rogue", mkmB1AssailantOracle, 3, 3)
	if !effectstest.TmtB1Card(t, g, id).Tapped {
		t.Error("Alley Assailant did not enter tapped")
	}
}

func TestMkmB1AlleyAssailantFaceDownEntersUntappedAndDrainsWhenTurnedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.MkmB1CastFaceDown(t, g, "Alley Assailant", "Creature — Vampire Rogue", mkmB1AssailantOracle, 3, 3)
	if effectstest.TmtB1Card(t, g, id).Tapped {
		t.Error("a face-down Assailant entered tapped; it has no abilities (CR 708.2a)")
	}
	myLife, theirLife := me.Life, opp.Life

	effectstest.MkmB1TurnFaceUp(t, g, me, id, "B", "B", "C", "C", "C", "C")
	mkmB1PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != theirLife-3 || me.Life != myLife+3 {
		t.Errorf("life: opponent %d -> %d, me %d -> %d; want -3 and +3", theirLife, opp.Life, myLife, me.Life)
	}
}

// --- Culvert Ambusher ----------------------------------------------------------

func TestMkmB1CulvertAmbusherEntersAndForcesABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Culvert Ambusher", "Creature — Wurm Horror", mkmB1AmbusherOracle, 4, 5)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.MkmB1MustBlock(t, g, bear) {
		t.Error("the target has no blocks-if-able requirement")
	}
}

func TestMkmB1CulvertAmbusherTriggersWhenTurnedFaceUpNotWhenEnteringFaceDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.MkmB1CastFaceDown(t, g, "Culvert Ambusher", "Creature — Wurm Horror", mkmB1AmbusherOracle, 4, 5)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Ambusher")
	}
	effectstest.MkmB1TurnFaceUp(t, g, me, id, "G", "C", "C", "C", "C")
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.MkmB1MustBlock(t, g, bear) {
		t.Error("turning the Ambusher face up put no blocks-if-able requirement on the target")
	}
}

// --- Benthic Criminologists ----------------------------------------------------

func TestMkmB1BenthicCriminologistsSacrificesAnArtifactToDrawOnEntering(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.CastHoldCreature(t, g, "Benthic Criminologists", "Creature — Merfolk Wizard", mkmB1BenthicOracle, 4, 5)
	hand := me.Hand.Size()

	effectstest.AnswerOwnPermanents(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

func TestMkmB1BenthicCriminologistsTriggersOnAttackAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	benthic := effectstest.B12Push(g, me.ID, "Benthic Criminologists", "Creature — Merfolk Wizard", mkmB1BenthicOracle, 4, 5)
	effectstest.DeclareAttack(t, g, opp.ID, benthic)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()

	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
		t.Fatal("attacking did not ask to sacrifice an artifact")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || me.Hand.Size() != hand {
		t.Error("declining sacrificed the artifact or drew a card")
	}
}

func TestMkmB1BenthicCriminologistsAsksNothingWithoutAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastHoldCreature(t, g, "Benthic Criminologists", "Creature — Merfolk Wizard", mkmB1BenthicOracle, 4, 5)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("asked to sacrifice with no artifact to sacrifice")
	}
}

// --- Cornered Crook ------------------------------------------------------------

func TestMkmB1CorneredCrookSacrificesAnArtifactForThreeToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.CastHoldCreature(t, g, "Cornered Crook", "Creature — Lizard Warrior", mkmB1CrookOracle, 5, 4)
	life := opp.Life

	effectstest.AnswerOwnPermanents(t, g, me.ID, rock)
	effectstest.B04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact was not sacrificed before the reflexive trigger asked for a target")
	}
	mkmB1PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
}

func TestMkmB1CorneredCrookHitsACreatureAndMayDecline(t *testing.T) {
	t.Run("creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
		wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
		effectstest.CastHoldCreature(t, g, "Cornered Crook", "Creature — Lizard Warrior", mkmB1CrookOracle, 5, 4)
		effectstest.AnswerOwnPermanents(t, g, me.ID, rock)
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickCard(t, g, me.ID, wall)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.TmtB1Card(t, g, wall).DamageMarked; got != 3 {
			t.Errorf("damage marked = %d, want 3", got)
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
		effectstest.CastHoldCreature(t, g, "Cornered Crook", "Creature — Lizard Warrior", mkmB1CrookOracle, 5, 4)
		effectstest.AnswerOwnPermanents(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(rock) || effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Error("declining sacrificed the artifact or raised the damage trigger")
		}
	})
}

// --- Case of the Crimson Pulse -------------------------------------------------

func TestMkmB1CrimsonPulseEntersDiscardingOneAndDrawingTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitched", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Case of the Crimson Pulse", "Enchantment — Case", mkmB1CrimsonPulseOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand, lib := me.Hand.Size(), me.Library.Size()

	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch) {
		t.Error("the chosen card was not discarded")
	}
	if got := lib - me.Library.Size(); got != 2 {
		t.Errorf("library fell by %d, want 2", got)
	}
	if got := me.Hand.Size(); got != hand-1+2 {
		t.Errorf("hand %d -> %d, want one discarded and two drawn", hand, got)
	}
}

func TestMkmB1CrimsonPulseSolvesOnlyWithAnEmptyHand(t *testing.T) {
	for _, empty := range []bool{true, false} {
		name := "holding cards"
		if empty {
			name = "empty hand"
		}
		t.Run(name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := effectstest.PushClass(g, me.ID, "Case of the Crimson Pulse", "Enchantment — Case", mkmB1CrimsonPulseOracl)
			effectstest.AdvanceToMain(t, g)
			if empty {
				g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, me.Hand.Size()) })
			}
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.MkmB1Solved(g, id) != empty {
				t.Errorf("solved = %v with %d cards in hand", effectstest.MkmB1Solved(g, id), me.Hand.Size())
			}
		})
	}
}

func TestMkmB1CrimsonPulseSolvedWheelsAtYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := effectstest.PushClass(g, me.ID, "Case of the Crimson Pulse", "Enchantment — Case", mkmB1CrimsonPulseOracl)
	g.WithWriteLock(func() {
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
	})
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	held := make([]uuid.UUID, 0, me.Hand.Size())
	for _, c := range me.Hand.Cards {
		held = append(held, c.InstanceID)
	}
	effectstest.PassPriorityAroundTable(t, g)

	for _, c := range held {
		if !me.Graveyard.Contains(c) {
			t.Errorf("card %s was not discarded", c)
		}
	}
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand is %d cards after the wheel, want 2", got)
	}
}

func TestMkmB1CrimsonPulseUnsolvedDoesNothingAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.PushClass(g, me.ID, "Case of the Crimson Pulse", "Enchantment — Case", mkmB1CrimsonPulseOracl)
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d at an unsolved Case's upkeep", hand, me.Hand.Size())
	}
}

// --- Case of the Filched Falcon ------------------------------------------------

func TestMkmB1FilchedFalconInvestigatesAndSolvesWithThreeArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Case of the Filched Falcon", "Enchantment — Case", mkmB1FalconOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Fatalf("%d Clues, want 1", got)
	}
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Case of the Filched Falcon" {
			id = c.InstanceID
		}
	}

	// The Clue and one more: two artifacts, not solved.
	effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.MkmB1Solved(g, id) {
		t.Fatal("the Case solved with two artifacts")
	}
	effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	for i := 0; i < 400 && !effectstest.MkmB1Solved(g, id); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !effectstest.MkmB1Solved(g, id) {
		t.Error("the Case never solved with three artifacts")
	}
}

func TestMkmB1FilchedFalconSolvedMakesANoncreatureArtifactABird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushClass(g, me.ID, "Case of the Filched Falcon", "Enchantment — Case", mkmB1FalconOracle)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 3, 3)

	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{2}{U}") })
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Fatal("an unsolved Case activated its solved ability")
	}
	g.WithWriteLock(func() {
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: golem}},
	}); err == nil {
		t.Fatal("the ability targeted an artifact creature")
	}
	effectstest.P7Activate(t, g, me, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	})

	if g.Battlefield.Contains(id) {
		t.Error("the Case was not sacrificed")
	}
	if got := effectstest.PlusOnes(g, rock); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4", got)
	}
	c := effectstest.LayeredCard(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Bird") {
		t.Errorf("the artifact is not an artifact Bird creature: %+v", c.Effective())
	}
	effectstest.AssertKeywords(t, g, rock, "flying")
	if p, tough := effectstest.MkmB1PT(t, g, rock); p != 4 || tough != 4 {
		t.Errorf("the Bird is %d/%d, want 4/4", p, tough)
	}
}
