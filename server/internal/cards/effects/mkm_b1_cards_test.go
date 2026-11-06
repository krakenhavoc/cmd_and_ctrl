package effects

import (
	"testing"

	"github.com/google/uuid"

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
		if ac.ManaCost != MorphCastCost || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != tc.faceUp {
			t.Errorf("%s: cast %q, kind %q, face-up %q; want %q, disguised, %q",
				tc.name, ac.ManaCost, ac.FaceDown.Kind, ac.FaceDown.FaceUpCost, MorphCastCost, tc.faceUp)
		}
	}
}

// --- shared helpers ------------------------------------------------------------

// mkmB1CastFaceDown casts a catalog card from the active player's hand
// face down for its disguise and resolves it.
func mkmB1CastFaceDown(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	advanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: "disguise"}); err != nil {
		t.Fatalf("cast %s face down: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if c := findCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatalf("%s did not enter face down", name)
	}
	return id
}

// mkmB1TurnFaceUp pays the disguise cost from conjured mana and turns
// the permanent face up.
func mkmB1TurnFaceUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, mana ...string) {
	t.Helper()
	g.WithWriteLock(func() { p.ManaPool.AddMana(manaTokens(mana...)...) })
	if err := g.PerformSpecialAction(p.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up: %v", err)
	}
	if findCardAnywhere(t, g, id).FaceDown {
		t.Fatal("still face down after the special action")
	}
}

// mkmB1PickPlayer answers the open pick_target prompt with a player.
func mkmB1PickPlayer(t *testing.T, g *game.Game, chooser, target uuid.UUID) {
	t.Helper()
	if err := tmcB7TryPickPlayer(t, g, chooser, target); err != nil {
		t.Fatalf("pick player: %v", err)
	}
}

// mkmB1PT is a battlefield permanent's power and toughness after the
// layers and its counters.
func mkmB1PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := tmtB1Card(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

// mkmB1MustBlock reports whether a battlefield creature carries a
// "blocks this turn if able" requirement.
func mkmB1MustBlock(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	for _, r := range c.Effective().BlockRequirements {
		if r.Kind == game.BlockRequirementBlocks {
			return true
		}
	}
	return false
}

// mkmB1Solved reads a Case's solved designation.
func mkmB1Solved(g *game.Game, id uuid.UUID) bool { return isSolved(g, id) }

// --- Auspicious Arrival --------------------------------------------------------

func TestMkmB1AuspiciousArrivalPumpsThenInvestigates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Auspicious Arrival", "Instant", mkmB1ArrivalOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)

	if p, tough := mkmB1PT(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the target is %d/%d, want 4/4", p, tough)
	}
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

// --- Cold Case Cracker ---------------------------------------------------------

func TestMkmB1ColdCaseCrackerFliesAndInvestigatesWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cracker := b12Push(g, me.ID, "Cold Case Cracker", "Creature — Spirit Detective", mkmB1CrackerOracle, 3, 3)
	assertKeywords(t, g, cracker, "flying")

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cracker) })
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues after the Cracker died, want 1", got)
	}
}

// --- Bolrac-Clan Basher --------------------------------------------------------

func TestMkmB1BolracClanBasherKeywordsAndDisguise(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmB1CastFaceDown(t, g, "Bolrac-Clan Basher", "Creature — Cyclops Warrior", mkmB1BasherOracle, 3, 2)
	if hasKeywordOnBattlefield(t, g, id, "double strike") || hasKeywordOnBattlefield(t, g, id, "trample") {
		t.Error("the face-down Basher kept a printed keyword (CR 708.2a)")
	}
	mkmB1TurnFaceUp(t, g, me, id, "R", "R", "C", "C", "C")
	assertKeywords(t, g, id, "double strike", "trample")
	if p, tough := mkmB1PT(t, g, id); p != 3 || tough != 2 {
		t.Errorf("face up the Basher is %d/%d, want 3/2", p, tough)
	}
}

// --- Basilica Stalker ----------------------------------------------------------

func TestMkmB1BasilicaStalkerGainsOneAndSurveilsOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stalker := b12Push(g, me.ID, "Basilica Stalker", "Creature — Vampire Detective", mkmB1StalkerOracle, 3, 4)
	assertKeywords(t, g, stalker, "flying")
	life := me.Life

	attackWith(t, g, opp.ID, stalker)
	passPriorityAroundTable(t, g)

	if got := me.Life; got != life+1 {
		t.Errorf("life %d -> %d, want +1", life, got)
	}
	c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
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
	g := newCatalogGame(t)
	id := castHoldCreature(t, g, "Alley Assailant", "Creature — Vampire Rogue", mkmB1AssailantOracle, 3, 3)
	if !tmtB1Card(t, g, id).Tapped {
		t.Error("Alley Assailant did not enter tapped")
	}
}

func TestMkmB1AlleyAssailantFaceDownEntersUntappedAndDrainsWhenTurnedUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := mkmB1CastFaceDown(t, g, "Alley Assailant", "Creature — Vampire Rogue", mkmB1AssailantOracle, 3, 3)
	if tmtB1Card(t, g, id).Tapped {
		t.Error("a face-down Assailant entered tapped; it has no abilities (CR 708.2a)")
	}
	myLife, theirLife := me.Life, opp.Life

	mkmB1TurnFaceUp(t, g, me, id, "B", "B", "C", "C", "C", "C")
	mkmB1PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != theirLife-3 || me.Life != myLife+3 {
		t.Errorf("life: opponent %d -> %d, me %d -> %d; want -3 and +3", theirLife, opp.Life, myLife, me.Life)
	}
}

// --- Culvert Ambusher ----------------------------------------------------------

func TestMkmB1CulvertAmbusherEntersAndForcesABlock(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Culvert Ambusher", "Creature — Wurm Horror", mkmB1AmbusherOracle, 4, 5)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !mkmB1MustBlock(t, g, bear) {
		t.Error("the target has no blocks-if-able requirement")
	}
}

func TestMkmB1CulvertAmbusherTriggersWhenTurnedFaceUpNotWhenEnteringFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := mkmB1CastFaceDown(t, g, "Culvert Ambusher", "Creature — Wurm Horror", mkmB1AmbusherOracle, 4, 5)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Ambusher")
	}
	mkmB1TurnFaceUp(t, g, me, id, "G", "C", "C", "C", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !mkmB1MustBlock(t, g, bear) {
		t.Error("turning the Ambusher face up put no blocks-if-able requirement on the target")
	}
}

// --- Benthic Criminologists ----------------------------------------------------

func TestMkmB1BenthicCriminologistsSacrificesAnArtifactToDrawOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	castHoldCreature(t, g, "Benthic Criminologists", "Creature — Merfolk Wizard", mkmB1BenthicOracle, 4, 5)
	hand := me.Hand.Size()

	answerOwnPermanents(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

func TestMkmB1BenthicCriminologistsTriggersOnAttackAndMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	benthic := b12Push(g, me.ID, "Benthic Criminologists", "Creature — Merfolk Wizard", mkmB1BenthicOracle, 4, 5)
	declareAttack(t, g, opp.ID, benthic)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()

	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) == nil {
		t.Fatal("attacking did not ask to sacrifice an artifact")
	}
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || me.Hand.Size() != hand {
		t.Error("declining sacrificed the artifact or drew a card")
	}
}

func TestMkmB1BenthicCriminologistsAsksNothingWithoutAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Benthic Criminologists", "Creature — Merfolk Wizard", mkmB1BenthicOracle, 4, 5)
	if latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("asked to sacrifice with no artifact to sacrifice")
	}
}

// --- Cornered Crook ------------------------------------------------------------

func TestMkmB1CorneredCrookSacrificesAnArtifactForThreeToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	castHoldCreature(t, g, "Cornered Crook", "Creature — Lizard Warrior", mkmB1CrookOracle, 5, 4)
	life := opp.Life

	answerOwnPermanents(t, g, me.ID, rock)
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact was not sacrificed before the reflexive trigger asked for a target")
	}
	mkmB1PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want -3", life, opp.Life)
	}
}

func TestMkmB1CorneredCrookHitsACreatureAndMayDecline(t *testing.T) {
	t.Run("creature", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
		wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 5)
		castHoldCreature(t, g, "Cornered Crook", "Creature — Lizard Warrior", mkmB1CrookOracle, 5, 4)
		answerOwnPermanents(t, g, me.ID, rock)
		b04WaitForPick(t, g, me.ID)
		pickCard(t, g, me.ID, wall)
		passPriorityAroundTable(t, g)
		if got := tmtB1Card(t, g, wall).DamageMarked; got != 3 {
			t.Errorf("damage marked = %d, want 3", got)
		}
	})
	t.Run("declined", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
		castHoldCreature(t, g, "Cornered Crook", "Creature — Lizard Warrior", mkmB1CrookOracle, 5, 4)
		answerOwnPermanents(t, g, me.ID)
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(rock) || latestPickTarget(g, me.ID) != nil {
			t.Error("declining sacrificed the artifact or raised the damage trigger")
		}
	})
}

// --- Case of the Crimson Pulse -------------------------------------------------

func TestMkmB1CrimsonPulseEntersDiscardingOneAndDrawingTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitched", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Case of the Crimson Pulse", "Enchantment — Case", mkmB1CrimsonPulseOracl, nil)
	passPriorityAroundTable(t, g)
	hand, lib := me.Hand.Size(), me.Library.Size()

	answerDiscard(t, g, me.ID, pitch)
	passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			id := pushClass(g, me.ID, "Case of the Crimson Pulse", "Enchantment — Case", mkmB1CrimsonPulseOracl)
			advanceToMain(t, g)
			if empty {
				g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, me.Hand.Size()) })
			}
			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
			if mkmB1Solved(g, id) != empty {
				t.Errorf("solved = %v with %d cards in hand", mkmB1Solved(g, id), me.Hand.Size())
			}
		})
	}
}

func TestMkmB1CrimsonPulseSolvedWheelsAtYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := pushClass(g, me.ID, "Case of the Crimson Pulse", "Enchantment — Case", mkmB1CrimsonPulseOracl)
	g.WithWriteLock(func() {
		if err := g.SolveCaseForEffect(id); err != nil {
			t.Fatalf("SolveCaseForEffect: %v", err)
		}
	})
	advanceToUpkeepOfSeat(t, g, seat)
	held := make([]uuid.UUID, 0, me.Hand.Size())
	for _, c := range me.Hand.Cards {
		held = append(held, c.InstanceID)
	}
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	pushClass(g, me.ID, "Case of the Crimson Pulse", "Enchantment — Case", mkmB1CrimsonPulseOracl)
	advanceToUpkeepOfSeat(t, g, seat)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d at an unsolved Case's upkeep", hand, me.Hand.Size())
	}
}

// --- Case of the Filched Falcon ------------------------------------------------

func TestMkmB1FilchedFalconInvestigatesAndSolvesWithThreeArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Case of the Filched Falcon", "Enchantment — Case", mkmB1FalconOracle, nil)
	passPriorityAroundTable(t, g)
	if got := tmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Fatalf("%d Clues, want 1", got)
	}
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Case of the Filched Falcon" {
			id = c.InstanceID
		}
	}

	// The Clue and one more: two artifacts, not solved.
	b12Permanent(g, me.ID, "Rock A", "Artifact")
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if mkmB1Solved(g, id) {
		t.Fatal("the Case solved with two artifacts")
	}
	b12Permanent(g, me.ID, "Rock B", "Artifact")
	for i := 0; i < 400 && !mkmB1Solved(g, id); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !mkmB1Solved(g, id) {
		t.Error("the Case never solved with three artifacts")
	}
}

func TestMkmB1FilchedFalconSolvedMakesANoncreatureArtifactABird(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := pushClass(g, me.ID, "Case of the Filched Falcon", "Enchantment — Case", mkmB1FalconOracle)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	golem := b12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 3, 3)

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
	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	})

	if g.Battlefield.Contains(id) {
		t.Error("the Case was not sacrificed")
	}
	if got := plusOnes(g, rock); got != 4 {
		t.Errorf("%d +1/+1 counters, want 4", got)
	}
	c := layeredCard(t, g, rock)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Bird") {
		t.Errorf("the artifact is not an artifact Bird creature: %+v", c.Effective())
	}
	assertKeywords(t, g, rock, "flying")
	if p, tough := mkmB1PT(t, g, rock); p != 4 || tough != 4 {
		t.Errorf("the Bird is %d/%d, want 4/4", p, tough)
	}
}
