package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// learn_cards_test.go — the cards that learn (#3525): each one's own
// clause, then the learn it ends with. The keyword action itself is
// learn_test.go.

const (
	academicDisputeOracle     = "b22de128-6dad-4e15-a547-1c9ca08df5ff"
	arcaneSubtractionOracle   = "4b5806fa-430f-4b78-b874-63d01a8a9de5"
	divideByZeroOracle        = "7b1c1a0c-f443-4eb5-a430-9eea886361ca"
	dreamStrixOracle          = "bfd6a73e-d126-4afe-99db-98cf716a1772"
	enthusiasticStudyOracle   = "4233334d-bb19-41e4-a203-2b7bb0c48f52"
	eyetwitchOracle           = "6a7a5e93-fea8-4aa0-a2b1-73d466220cfb"
	fieldTripOracle           = "29423b70-7cb7-4617-898c-78a38eb233ad"
	firstDayOfClassOracle     = "e0b5c772-e029-4b8d-be67-b1584700e377"
	gnarledProfessorOracle    = "e66083f0-62a9-4ff7-8bb0-3eafc6d2e95d"
	guidingVoiceOracle        = "5c7e7c9e-00d1-497e-bee7-278bbd82e234"
	huntForSpecimensOracle    = "53f1bcff-55fb-4d38-b108-dcbed623f488"
	igneousInspirationOracle  = "2e31af8f-5e94-4888-857d-911253eac2cd"
	overgrownArchOracle       = "802f8bcd-9219-48cd-a4c5-d9ea9ef60c9e"
	poetsQuillOracle          = "f474d421-5fd0-4d76-95cb-4824a2381a29"
	popQuizOracle             = "dda9ac6c-1ab2-4f77-b499-f03d587163d8"
	professorOfSymbologyOracl = "1a8ae444-6237-4591-97ab-f43e181b2f4f"
	riseOfExtusOracle         = "810f2b50-59c7-490a-8e6a-00f982cd16a0"
	sparringRegimenOracle     = "2aac3575-4a06-452c-bb0f-68dc34086226"
	studyBreakOracle          = "7aae7c7b-0b31-4b48-a8d3-81c913f604cf"
)

// learnCardTable is a two-seat game whose seat 0 owns one Lesson outside
// the game.
func learnCardTable(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := newSideboardCatalogGame(t, 2, lessonCard("Environmental Sciences"))
	return g, g.Seats[0], g.Seats[1]
}

// takesTheLesson answers the learn that should now be open: no discard,
// then the Lesson, which must come into the hand.
func takesTheLesson(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	c := learnDiscardPrompt(g, me)
	if c == nil {
		t.Fatalf("learn did not ask what to discard: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline the discard: %v", err)
	}
	passPriorityAroundTable(t, g)
	lc := learnLessonPrompt(g, me)
	if lc == nil {
		t.Fatalf("learn did not offer the Lesson: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{sideCardID(me, "Environmental Sciences")}); err != nil {
		t.Fatalf("take the Lesson: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Environmental Sciences") {
		t.Fatal("the Lesson did not come into the hand")
	}
}

// noLearnYet fails when learn has already asked: a step that has to wait
// for an earlier answer must not have run.
func noLearnYet(t *testing.T, g *game.Game, me *game.Player, why string) {
	t.Helper()
	if learnDiscardPrompt(g, me) != nil || learnLessonPrompt(g, me) != nil {
		t.Fatalf("learn asked before %s: %+v", why, g.PendingChoices)
	}
}

// answerConfirmFor answers the open yes/no a resolving effect asked.
func answerConfirmFor(t *testing.T, g *game.Game, me *game.Player, yes bool) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == me.ID {
			if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			return
		}
	}
	t.Fatalf("no yes/no prompt for %s: %+v", me.Name, g.PendingChoices)
}

func castTargeting(t *testing.T, g *game.Game, name, typeLine, oracle string, ids ...uuid.UUID) {
	t.Helper()
	var targets []game.TargetRef
	for _, id := range ids {
		targets = append(targets, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	castCatalogSpell(t, g, name, typeLine, oracle, targets)
	passPriorityAroundTable(t, g)
}

// --- spells ------------------------------------------------------------

func TestPopQuizDrawsThenLearns(t *testing.T) {
	g, me, _ := learnCardTable(t)
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Pop Quiz", "Instant", popQuizOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d, want the draw", before, me.Hand.Size())
	}
	takesTheLesson(t, g, me)
}

func TestHuntForSpecimensMakesAPestThenLearns(t *testing.T) {
	g, me, _ := learnCardTable(t)
	castCatalogSpell(t, g, "Hunt for Specimens", "Sorcery", huntForSpecimensOracle, nil)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Pest") == uuid.Nil {
		t.Error("no Pest token")
	}
	takesTheLesson(t, g, me)
}

// The search asks first; learn waits for the Forest to be chosen.
func TestFieldTripFetchesATappedForestBeforeLearning(t *testing.T) {
	g, me, _ := learnCardTable(t)
	ids := seedSearchLibrary(me,
		searchTestLand("Plains", "Basic Land — Plains"),
		searchTestLand("Forest", "Basic Land — Forest"))
	castCatalogSpell(t, g, "Field Trip", "Sorcery", fieldTripOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[1] {
		t.Errorf("offered %v, want only the Forest %v", c.SearchCards, ids[1])
	}
	noLearnYet(t, g, me, "the Forest was chosen")
	answerSearchByID(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if card, ok := battlefieldCard(g, ids[1]); !ok || !card.Tapped {
		t.Errorf("the Forest should be on the battlefield tapped: %+v ok=%v", card, ok)
	}
	takesTheLesson(t, g, me)
}

func TestGuidingVoicePutsACounterThenLearns(t *testing.T) {
	g, me, _ := learnCardTable(t)
	bear := edCard(t, g, me, "Bear", "", "Creature — Bear", 2, 2)
	castTargeting(t, g, "Guiding Voice", "Sorcery", guidingVoiceOracle, bear)
	if n := plusOneCounters(g, bear); n != 1 {
		t.Errorf("+1/+1 counters %d, want 1", n)
	}
	takesTheLesson(t, g, me)
}

func TestIgneousInspirationDamagesTheTargetThenLearns(t *testing.T) {
	g, me, opp := learnCardTable(t)
	castCatalogSpell(t, g, "Igneous Inspiration", "Sorcery", igneousInspirationOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != 37 {
		t.Errorf("opponent life %d, want 37", opp.Life)
	}
	takesTheLesson(t, g, me)
}

// "Up to two": two creatures tap, and one or none is just as castable.
func TestStudyBreakTapsUpToTwoCreaturesThenLearns(t *testing.T) {
	g, me, opp := learnCardTable(t)
	a := edCard(t, g, opp, "Bear A", "", "Creature — Bear", 2, 2)
	b := edCard(t, g, opp, "Bear B", "", "Creature — Bear", 2, 2)
	castTargeting(t, g, "Study Break", "Instant", studyBreakOracle, a, b)
	if !battlefieldTapped(g, a) || !battlefieldTapped(g, b) {
		t.Errorf("tapped A %v, B %v, want both", battlefieldTapped(g, a), battlefieldTapped(g, b))
	}
	takesTheLesson(t, g, me)
}

func TestStudyBreakWithNoTargetStillLearns(t *testing.T) {
	g, me, _ := learnCardTable(t)
	castCatalogSpell(t, g, "Study Break", "Instant", studyBreakOracle, nil)
	passPriorityAroundTable(t, g)
	takesTheLesson(t, g, me)
}

func TestArcaneSubtractionShrinksPowerThenLearns(t *testing.T) {
	g, me, opp := learnCardTable(t)
	bear := edCard(t, g, opp, "Bear", "", "Creature — Bear", 5, 3)
	castTargeting(t, g, "Arcane Subtraction", "Instant", arcaneSubtractionOracle, bear)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 1 || tough != 3 {
		t.Errorf("%d/%d, want 1/3", p, tough)
	}
	takesTheLesson(t, g, me)
}

func TestEnthusiasticStudyPumpsAndGrantsTrampleThenLearns(t *testing.T) {
	g, me, _ := learnCardTable(t)
	bear := edCard(t, g, me, "Bear", "", "Creature — Bear", 2, 2)
	castTargeting(t, g, "Enthusiastic Study", "Instant", enthusiasticStudyOracle, bear)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("%d/%d, want 5/3", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("the creature did not gain trample")
	}
	takesTheLesson(t, g, me)
}

// "You may have it gain reach": asked after the block requirement, and
// learn waits for the answer.
func TestAcademicDisputeBlockRequirementOptionalReachThenLearns(t *testing.T) {
	for _, yes := range []bool{true, false} {
		name := "declined"
		if yes {
			name = "accepted"
		}
		t.Run(name, func(t *testing.T) {
			g, me, opp := learnCardTable(t)
			bear := edCard(t, g, opp, "Bear", "", "Creature — Bear", 2, 2)
			castTargeting(t, g, "Academic Dispute", "Instant", academicDisputeOracle, bear)
			required := false
			for _, e := range g.ScopedEffects {
				for _, m := range e.Mods {
					if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementBlocks) {
						required = true
					}
				}
			}
			if !required {
				t.Error("no blocks-if-able requirement on the target")
			}
			noLearnYet(t, g, me, "the reach question was answered")
			answerConfirmFor(t, g, me, yes)
			passPriorityAroundTable(t, g)
			if got := effectiveAbilitiesContain(t, g, bear, "reach"); got != yes {
				t.Errorf("reach = %v, want %v", got, yes)
			}
			takesTheLesson(t, g, me)
		})
	}
}

// Two clauses with a target each, then learn. The graveyard card is
// optional: "up to one".
func TestRiseOfExtusExilesBothTargetsThenLearns(t *testing.T) {
	g, me, opp := learnCardTable(t)
	bear := edCard(t, g, opp, "Bear", "", "Creature — Bear", 2, 2)
	bolt := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Shock", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Rise of Extus", "Sorcery", riseOfExtusOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0},
		{Kind: game.TargetCard, ID: bolt, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Bear") != uuid.Nil || !outsideExileHas(g, "Bear") {
		t.Error("the creature was not exiled")
	}
	if opp.Graveyard.Contains(bolt) || !outsideExileHas(g, "Shock") {
		t.Error("the instant was not exiled from the graveyard")
	}
	takesTheLesson(t, g, me)
}

func TestRiseOfExtusWithOnlyACreatureStillLearns(t *testing.T) {
	g, me, opp := learnCardTable(t)
	bear := edCard(t, g, opp, "Bear", "", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Rise of Extus", "Sorcery", riseOfExtusOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0},
	})
	passPriorityAroundTable(t, g)
	if !outsideExileHas(g, "Bear") {
		t.Error("the creature was not exiled")
	}
	takesTheLesson(t, g, me)
}

// Mana value 1 or greater: a permanent with a cost can be returned, one
// with none (a token, a free permanent) cannot be targeted.
func TestDivideByZeroReturnsAPermanentWithAManaValueThenLearns(t *testing.T) {
	g, me, opp := learnCardTable(t)
	costly := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Two Drop", TypeLine: "Artifact", ManaCost: "{2}"})
	free := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Free Thing", TypeLine: "Artifact"})
	if err := castCatalogSpellErr(t, g, "Divide by Zero", "Instant", divideByZeroOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: free}}); err == nil {
		t.Fatal("a permanent with mana value 0 was a legal target")
	}
	castTargeting(t, g, "Divide by Zero", "Instant", divideByZeroOracle, costly)
	if findBattlefieldCardForTest(g, costly) != nil || !handHasID(opp, costly) {
		t.Error("the permanent was not returned to its owner's hand")
	}
	takesTheLesson(t, g, me)
}

// --- creatures and permanents --------------------------------------------

func TestEntersLearnersAskWhenTheyEnter(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Gnarled Professor", "Creature — Treefolk Druid", gnarledProfessorOracle},
		{"Professor of Symbology", "Creature — Kor Cleric", professorOfSymbologyOracl},
		{"Poet's Quill", "Artifact — Equipment", poetsQuillOracle},
		{"Sparring Regimen", "Enchantment", sparringRegimenOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := learnCardTable(t)
			castCreatureCard(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			takesTheLesson(t, g, me)
		})
	}
}

func TestDiesLearnersAskWhenTheyDie(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Eyetwitch", eyetwitchOracle},
		{"Dream Strix", dreamStrixOracle},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := learnCardTable(t)
			dier := apaPush(g, me.ID, me.ID, game.Card{Name: tc.name, OracleID: tc.oracle, TypeLine: "Creature — Test", Power: 1, Toughness: 1})
			noLearnYet(t, g, me, "it died")
			g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dier) })
			passPriorityAroundTable(t, g)
			takesTheLesson(t, g, me)
		})
	}
}

// A spell that targets it makes it sacrifice itself, and the sacrifice is
// a death, so it learns.
func TestDreamStrixIsSacrificedWhenASpellTargetsItAndThenLearns(t *testing.T) {
	g, me, opp := learnCardTable(t)
	strix := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Dream Strix", OracleID: dreamStrixOracle, TypeLine: "Creature — Bird Illusion", Power: 3, Toughness: 2})
	castTargeting(t, g, "Guiding Voice", "Sorcery", guidingVoiceOracle, strix)
	if findBattlefieldCardForTest(g, strix) != nil {
		t.Fatal("the Strix survived being targeted by a spell")
	}
	if !opp.Graveyard.Contains(strix) {
		t.Fatal("the Strix is not in its owner's graveyard")
	}
	// Learn belongs to the Strix's controller, not to the spell's.
	if learnDiscardPrompt(g, opp) == nil {
		t.Fatalf("the Strix's controller was not asked: %+v", g.PendingChoices)
	}
	if learnDiscardPrompt(g, me) != nil {
		t.Error("the spell's controller learned")
	}
}

// An ability's target is not a spell's: the Strix stays.
func TestDreamStrixSurvivesBeingTargetedByAnAbility(t *testing.T) {
	spec, ok := Lookup(dreamStrixOracle)
	if !ok || len(spec.Triggered) != 2 {
		t.Fatalf("Dream Strix has %d triggers, want the sacrifice and the learn", len(spec.Triggered))
	}
	if spec.Triggered[0].AppliesTo == nil {
		t.Fatal("the sacrifice trigger has no condition")
	}
}

func TestOvergrownArchSacrificesItselfToLearn(t *testing.T) {
	g, me, _ := learnCardTable(t)
	arch := apaPush(g, me.ID, me.ID, game.Card{Name: "Overgrown Arch", OracleID: overgrownArchOracle, TypeLine: "Creature — Plant Wall", Power: 0, Toughness: 4})

	// {T}: you gain 1 life.
	life := me.Life
	p7Activate(t, g, me, arch, 0, game.ActivateAbilityParams{})
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}

	// {2}, sacrifice: learn.
	p7Activate(t, g, me, arch, 1, game.ActivateAbilityParams{})
	if findBattlefieldCardForTest(g, arch) != nil || !me.Graveyard.Contains(arch) {
		t.Error("the Arch was not sacrificed")
	}
	takesTheLesson(t, g, me)
}

// Poet's Quill: the equipped creature gets +1/+1 and lifelink.
func TestPoetsQuillEquipsForPlusOneAndLifelink(t *testing.T) {
	g, me, _ := learnCardTable(t)
	advanceToMain(t, g)
	bear := edCard(t, g, me, "Bear", "", "Creature — Bear", 2, 2)
	quill := apaPush(g, me.ID, me.ID, game.Card{Name: "Poet's Quill", OracleID: poetsQuillOracle, TypeLine: "Artifact — Equipment"})
	equipTo(t, g, me.ID, quill, bear)
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("%d/%d, want 3/3", p, tough)
	}
	if !effectiveAbilitiesContain(t, g, bear, "lifelink") {
		t.Error("the equipped creature lacks lifelink")
	}
}

// Sparring Regimen: once per attack, a counter on the target attacker
// and untap it.
func TestSparringRegimenPutsACounterOnAnAttackerAndUntapsIt(t *testing.T) {
	g, me, opp := learnCardTable(t)
	apaPush(g, me.ID, me.ID, game.Card{Name: "Sparring Regimen", OracleID: sparringRegimenOracle, TypeLine: "Enchantment"})
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("first attack: %v", err)
	}
	if err := g.DeclareAttacker(b, opp.ID); err != nil {
		t.Fatalf("second attack: %v", err)
	}
	lockInAttacks(t, g)
	answerPickTarget(t, g, a)
	passPriorityAroundTable(t, g)
	if plusOneCounters(g, a) != 1 || plusOneCounters(g, b) != 0 {
		t.Errorf("counters A %d, B %d, want 1 and 0", plusOneCounters(g, a), plusOneCounters(g, b))
	}
	if battlefieldTapped(g, a) {
		t.Error("the targeted attacker is still tapped")
	}
	if !battlefieldTapped(g, b) {
		t.Error("the other attacker was untapped")
	}
}

// --- First Day of Class ------------------------------------------------

// Until end of turn, each creature that enters under your control gets
// a counter and haste; one that entered before does not.
func TestFirstDayOfClassBoostsCreaturesThatEnterThisTurn(t *testing.T) {
	g, me, opp := learnCardTable(t)
	early := edCard(t, g, me, "Early Bear", "", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "First Day of Class", "Instant", firstDayOfClassOracle, nil)
	passPriorityAroundTable(t, g)
	takesTheLesson(t, g, me)

	first := castCreatureCard(t, g, "Late Bear", "Creature — Bear", "", 2, 2)
	second := castCreatureCard(t, g, "Later Bear", "Creature — Bear", "", 3, 3)
	if n := plusOneCounters(g, early); n != 0 {
		t.Errorf("a creature that entered before has %d counters", n)
	}
	for _, id := range []uuid.UUID{first, second} {
		if n := plusOneCounters(g, id); n != 1 {
			t.Errorf("%s has %d +1/+1 counters, want 1", id, n)
		}
		if !effectiveAbilitiesContain(t, g, id, "haste") {
			t.Errorf("%s did not gain haste", id)
		}
	}
	// "A creature you control": an opponent's creature entering is not
	// affected.
	theirs := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: theirs, Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	})
	g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Token", TypeLine: "Creature — Bear", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	passPriorityAroundTable(t, g)
}
