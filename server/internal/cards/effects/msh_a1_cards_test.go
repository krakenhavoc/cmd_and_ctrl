package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a1_cards_test.go — Marvel Super Heroes, half A, first batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA1AimScientistsOracle = "fa732888-5b7d-4087-a6e5-1db1c98fbd4c"
	mshA1AgentOfAtlasOracle  = "cd403866-bef1-413a-85e2-c86e295b3589"
	mshA1AgentsShieldOracle  = "203b72d2-a273-49cc-bd10-2fbc07ed2fc8"
	mshA1AntManArmyOracle    = "34c0f75d-10a9-428d-b202-3defb3949090"
	mshA1ArmorWarsOracle     = "1ad83cac-2c85-477b-ab00-1369f9683db5"
	mshA1AtlanteanOracle     = "e929b730-9aa2-49e1-9a65-50afc0fbe3e0"
	mshA1AssembleOracle      = "3477e1cc-46af-46f4-9ea2-bdf0bf1f880e"
	mshA1TowerOracle         = "c5fc8e7c-a87e-4586-a13c-d30e0a3aafbf"
	mshA1BaxterOracle        = "71bc69a5-7cec-4abd-b97d-13f8e1f9afac"
	mshA1BlackPantherOracle  = "b0bf1964-c543-4c53-9f09-0dd7c608d5c5"
	mshA1DailyBugleOracle    = "d2a3f735-9523-4ddd-96a9-b0a5ecd39284"
	mshA1FeralFerocityOracle = "5e968bf3-8d83-480b-81a4-98883cafd04f"
	mshA1WebUpOracle         = "9b9b086c-d201-4574-a1e9-0d6f8de53d7d"
)

// mshA1GainLands are the ten Marvel printings of the gain-land cycle.
var mshA1GainLands = []struct{ name, oracle, a, b string }{
	{"Subterranean Cavern", "d49a7525-eca5-48e1-b94b-d618a706fd02", "B", "G"},
	{"Pym Technologies", "5a8977fc-0732-4807-b876-01819aedd70b", "G", "U"},
	{"Los Diablos Missile Base", "4b17048f-8c7f-4e29-8f56-16db4c2106e6", "R", "G"},
	{"Birnin Zana Plaza", "c3172c5d-c745-4422-b651-cd414c8da11b", "G", "W"},
	{"Hell's Kitchen", "1337d89e-a883-4568-a14f-9ca3305cc3df", "B", "R"},
	{"Fisk Tower", "5554db95-9676-4c24-b82a-8173513d7927", "W", "B"},
	{"Avengers Hangar", "10fc2009-4212-47c6-8a97-ab0f45c1f91b", "W", "U"},
	{"A.I.M. Labs", "8417f2d5-93c9-474a-a143-5c4e17097a88", "U", "B"},
	{"Asgardian Citadel", "e3614df6-ce84-4180-ade1-a1751e8e3c9a", "R", "W"},
	{"Stark Industries", "52401cb4-1bce-4191-9220-0d93ea4108b8", "U", "R"},
}

func TestMshA1EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA1AimScientistsOracle: "A.I.M. Scientists",
		mshA1AgentOfAtlasOracle:  "Agent of Atlas",
		mshA1AgentsShieldOracle:  "Agents of S.H.I.E.L.D.",
		mshA1AntManArmyOracle:    "Ant-Man's Army",
		mshA1ArmorWarsOracle:     "Armor Wars",
		mshA1AtlanteanOracle:     "Atlantean Cavalry",
		mshA1AssembleOracle:      "Avengers Assemble!",
		mshA1TowerOracle:         "Avengers Tower",
		mshA1BaxterOracle:        "Baxter Building",
		mshA1BlackPantherOracle:  "Black Panther, Vanguard",
		mshA1DailyBugleOracle:    "Daily Bugle Newspaper",
		mshA1FeralFerocityOracle: "Feral Ferocity",
		mshA1WebUpOracle:         "Web Up",
	}
	for _, l := range mshA1GainLands {
		want[l.oracle] = l.name
	}
	for oracle, name := range want {
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

// --- the ten gain lands --------------------------------------------------------

func TestMshA1GainLandsEnterTappedGainOneAndTapForEitherColour(t *testing.T) {
	for _, tc := range mshA1GainLands {
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
			t.Errorf("%s: life %d -> %d, want +1", tc.name, before, me.Life)
		}
		spec, _ := Lookup(tc.oracle)
		if want := "{" + tc.a + "|" + tc.b + "}"; len(spec.ManaAbilities) != 1 || spec.ManaAbilities[0].Produced != want {
			t.Errorf("%s: mana abilities %+v, want one producing %s", tc.name, spec.ManaAbilities, want)
		}
	}
}

func TestMshA1GainLandTapsForItsSecondColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := b12PlayFromHand(t, g, "Hell's Kitchen", "Land", mshA1GainLands[4].oracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b28TapForMana(t, g, me.ID, land, "R")
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("tapped for {R}: pool %v", got)
	}
}

// --- Daily Bugle Newspaper ---------------------------------------------------------

func TestMshA1DailyBugleNewspaperLootsAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paper := b12Push(g, me.ID, "Daily Bugle Newspaper", "Artifact", mshA1DailyBugleOracle, 0, 0)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, paper, 0, game.ActivateAbilityParams{})
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("drew %d, want 1", got-hand)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Errorf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
	if b16CountNamed(g, "Treasure") != 0 {
		t.Error("the Treasure was created with the discard still open")
	}
	discardFromHand(t, g, me.ID)
	if b16CountNamed(g, "Treasure") != 1 {
		t.Error("a Treasure is created")
	}
}

// --- Feral Ferocity ------------------------------------------------------------------

func TestMshA1FeralFerocityGivesPlusFourPlusFourToAnyCreatureUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Feral Ferocity", "Instant", mshA1FeralFerocityOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 6 {
		t.Errorf("power %d, want 6 (an opponent's creature is a legal target)", got)
	}
	advanceToNextSeatsTurn(t, g)
	if got := hobA4Power(t, g, bear); got != 2 {
		t.Errorf("power %d after the turn, want 2", got)
	}
}

// --- Web Up -------------------------------------------------------------------------------

func TestMshA1WebUpExilesUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	web := castCatalogSpell(t, g, "Web Up", "Enchantment", mshA1WebUpOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(web) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after Web Up left, %d copies under its owner's control, want 1", n)
	}
}

func TestMshA1WebUpOnlyTargetsOpponentsNonlandPermanents(t *testing.T) {
	spec, _ := Lookup(mshA1WebUpOracle)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirLand := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirLand: false, theirs: true})
}

// --- A.I.M. Scientists ---------------------------------------------------------------

const mshA1ScientistsType = "Creature — Human Scientist Villain"

func TestMshA1AimScientistsConniveGrowsOnANonlandDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := handCardForTest(me, "Spare Spell", "Sorcery", "")
	before := me.Hand.Size()
	id := castHoldCreature(t, g, "A.I.M. Scientists", mshA1ScientistsType, mshA1AimScientistsOracle, 3, 3)
	// The cast took the Scientists out of the hand and the connive drew one.
	if got := me.Hand.Size(); got != before+1 {
		t.Fatalf("hand %d, want %d (the draw, before the discard)", got, before+1)
	}
	if discardOwed(g, me.ID) != 1 {
		t.Fatalf("discard owed = %d, want 1", discardOwed(g, me.ID))
	}
	answerDiscard(t, g, me.ID, spell)
	if n := hobA8Counters(t, g, id); n != 1 {
		t.Errorf("%d +1/+1 counters after discarding a nonland card, want 1", n)
	}
}

func TestMshA1AimScientistsConniveGrowsNotOnALandDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := handCardForTest(me, "Spare Forest", "Basic Land — Forest", "")
	id := castHoldCreature(t, g, "A.I.M. Scientists", mshA1ScientistsType, mshA1AimScientistsOracle, 3, 3)
	answerDiscard(t, g, me.ID, forest)
	if n := hobA8Counters(t, g, id); n != 0 {
		t.Errorf("%d +1/+1 counters after discarding a land, want 0", n)
	}
}

func TestMshA1AimScientistsBasicLandcycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, "A.I.M. Scientists", mshA1ScientistsType, mshA1AimScientistsOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the landcycling search did not ask")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Fatal("no basic land reached the hand")
	}
}

// --- Agent of Atlas -------------------------------------------------------------------

func TestMshA1AgentOfAtlasHasProwess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	agent := pushCatalogPermanent(g, me.ID, "Agent of Atlas", "Creature — Human Spy Hero", mshA1AgentOfAtlasOracle, false)
	castNoncreature(t, g)
	settleProwess(t, g)
	if p, tg := effectivePower(t, g, agent), effectiveToughness(t, g, agent); p != 2 || tg != 2 {
		t.Errorf("after a noncreature spell: %d/%d, want 2/2 from a 1/1 body", p, tg)
	}
}

// --- Agents of S.H.I.E.L.D. ------------------------------------------------------------

func TestMshA1AgentsOfShieldPumpsTheCreatureThatAttacksAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Agents of S.H.I.E.L.D.", "Creature — Human Spy Hero", mshA1AgentsShieldOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 3 {
		t.Errorf("the lone attacker has power %d, want 3", got)
	}
}

func TestMshA1AgentsOfShieldDoNothingWhenTwoCreaturesAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Agents of S.H.I.E.L.D.", "Creature — Human Spy Hero", mshA1AgentsShieldOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	declareAttack(t, g, opp.ID, bear, wolf)
	passPriorityAroundTable(t, g)
	if a, b := hobA4Power(t, g, bear), hobA4Power(t, g, wolf); a != 2 || b != 2 {
		t.Errorf("powers %d and %d, want 2 and 2 (no creature attacked alone)", a, b)
	}
}

func TestMshA1AgentsOfShieldIgnoreAnOpponentsLoneAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Agents of S.H.I.E.L.D.", "Creature — Human Spy Hero", mshA1AgentsShieldOracle, false)
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2: the opponent's Agents only watch their own creatures", got)
	}
}

// --- Ant-Man's Army --------------------------------------------------------------------

func TestMshA1AntMansArmyMakesATreasureOrAFood(t *testing.T) {
	for _, tc := range []struct {
		pick          int
		token, absent string
	}{
		{0, "Treasure", "Food"},
		{1, "Food", "Treasure"},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		castHoldCreature(t, g, "Ant-Man's Army", "Creature — Insect", mshA1AntManArmyOracle, 3, 2)
		c := latestOptionPickFor(g, me.ID)
		if c == nil || len(c.PickOptions) != 2 {
			t.Fatalf("no Food-or-Treasure choice: %+v", c)
		}
		answerOptionPick(t, g, me.ID, tc.pick)
		if countBattlefieldNamed(g, me.ID, tc.token) != 1 || countBattlefieldNamed(g, me.ID, tc.absent) != 0 {
			t.Errorf("pick %d: want exactly one %s and no %s", tc.pick, tc.token, tc.absent)
		}
	}
}

// --- Atlantean Cavalry -----------------------------------------------------------------

func TestMshA1AtlanteanCavalryGrowsOnTheSecondDrawOnlyAndHasVigilance(t *testing.T) {
	hobA5SecondDrawGrows(t, "Atlantean Cavalry", "Creature — Merfolk Soldier", mshA1AtlanteanOracle)
	g := newCatalogGame(t)
	vig := pushCatalogPermanent(g, g.Seats[0].ID, "Atlantean Cavalry", "Creature — Merfolk Soldier", mshA1AtlanteanOracle, false)
	if !hasKeywordOnBattlefield(t, g, vig, "vigilance") {
		t.Error("Atlantean Cavalry has no vigilance")
	}
}

// --- Armor Wars -----------------------------------------------------------------------

// mshA1ArmorWarsTable seeds `artifacts` artifacts for seat 0 and casts the
// Saga, which settles up to its chapter I question.
func mshA1ArmorWarsTable(t *testing.T, artifacts int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < artifacts; i++ {
		b27Push(g, me.ID, "Test Rock", "Artifact", "", "{2}", 0, 0)
	}
	saga := castCatalogSpell(t, g, "Armor Wars", "Enchantment — Saga", mshA1ArmorWarsOracle, nil)
	passPriorityAroundTable(t, g)
	return g, saga
}

func mshA1AnswerConfirm(t *testing.T, g *game.Game, who uuid.UUID, yes bool) {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, who)
	if c == nil {
		t.Fatalf("no yes/no question for %s: %+v", who, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, who, yes); err != nil {
		t.Fatalf("ResolveConfirm(%v): %v", yes, err)
	}
}

func TestMshA1ArmorWarsDrawsOnePerArtifactAndEachOpponentDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Test Rock", "Artifact", "", "{2}", 0, 0)
	b27Push(g, me.ID, "Test Gear", "Artifact", "", "{1}", 0, 0)
	b27Push(g, g.Seats[1].ID, "Their Rock", "Artifact", "", "{4}", 0, 0)
	mine := me.Hand.Size()
	theirs := []int{g.Seats[1].Hand.Size(), g.Seats[2].Hand.Size(), g.Seats[3].Hand.Size()}
	castCatalogSpell(t, g, "Armor Wars", "Enchantment — Saga", mshA1ArmorWarsOracle, nil)
	passPriorityAroundTable(t, g)
	mshA1AnswerConfirm(t, g, me.ID, true)
	if got := me.Hand.Size() - mine; got != 2 {
		t.Errorf("drew %d, want 2 (my two artifacts, not the table's three)", got)
	}
	for i, s := range g.Seats[1:] {
		if got := s.Hand.Size() - theirs[i]; got != 1 {
			t.Errorf("opponent %d drew %d, want exactly 1", i+1, got)
		}
	}
}

func TestMshA1ArmorWarsDecliningDrawsNothingForAnyone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Test Rock", "Artifact", "", "{2}", 0, 0)
	mine, theirs := me.Hand.Size(), g.Seats[1].Hand.Size()
	castCatalogSpell(t, g, "Armor Wars", "Enchantment — Saga", mshA1ArmorWarsOracle, nil)
	passPriorityAroundTable(t, g)
	mshA1AnswerConfirm(t, g, me.ID, false)
	if me.Hand.Size() != mine || g.Seats[1].Hand.Size() != theirs {
		t.Errorf("hands %d and %d, want %d and %d", me.Hand.Size(), g.Seats[1].Hand.Size(), mine, theirs)
	}
}

func TestMshA1ArmorWarsWithNoArtifactAsksNothingAndDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine, theirs := me.Hand.Size(), g.Seats[1].Hand.Size()
	castCatalogSpell(t, g, "Armor Wars", "Enchantment — Saga", mshA1ArmorWarsOracle, nil)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil {
		t.Error("asked to draw zero cards")
	}
	if me.Hand.Size() != mine || g.Seats[1].Hand.Size() != theirs {
		t.Errorf("hands %d and %d, want %d and %d", me.Hand.Size(), g.Seats[1].Hand.Size(), mine, theirs)
	}
}

func TestMshA1ArmorWarsChapterTwoDiscountsArtifactSpellsForTheTurn(t *testing.T) {
	g, _ := mshA1ArmorWarsTable(t, 0)
	me := g.Seats[0]
	hobA9NextChapter(t, g)
	for range 2 {
		if got := priceInHand(t, g, me, "Test Rock", "Artifact", "{3}"); got != 2 {
			t.Errorf("an artifact spell costs %d, want 2 (every one this turn, not just the next)", got)
		}
	}
	if got := priceInHand(t, g, me, "Test Golem", "Artifact Creature — Golem", "{3}"); got != 2 {
		t.Errorf("an artifact creature spell costs %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Test Bear", "Creature — Bear", "{3}"); got != 3 {
		t.Errorf("a nonartifact creature spell costs %d, want 3", got)
	}
	if got := priceInHand(t, g, me, "Test Rite", "Sorcery", "{3}"); got != 3 {
		t.Errorf("a sorcery costs %d, want 3", got)
	}
	advanceToNextSeatsTurn(t, g)
	if got := priceInHand(t, g, me, "Test Rock", "Artifact", "{3}"); got != 3 {
		t.Errorf("an artifact spell costs %d on a later turn, want 3", got)
	}
}

func TestMshA1ArmorWarsChapterThreeDealsTheGreatestArtifactManaValueToAnOpponent(t *testing.T) {
	g, saga := mshA1ArmorWarsTable(t, 0)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Signet", "Artifact", "", "{2}", 0, 0)
	b27Push(g, me.ID, "Wurmcoil", "Artifact Creature — Wurm", "", "{6}", 6, 6)
	b27Push(g, opp.ID, "Their Colossus", "Artifact Creature — Golem", "", "{11}", 11, 11)
	hobA9NextChapter(t, g)
	hobA9NextChapter(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("chapter III asked for no target")
	}
	if hasID(p.PickTargetPlayers, me.ID) || !hasID(p.PickTargetPlayers, opp.ID) {
		t.Errorf("legal players %v: an opponent, never me", p.PickTargetPlayers)
	}
	life := opp.Life
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 6 {
		t.Errorf("dealt %d, want 6 (my greatest artifact, not the table's)", got)
	}
	if !inGraveyardOf(g, me.ID, saga) {
		t.Error("the Saga should be sacrificed after chapter III")
	}
}

func TestMshA1ArmorWarsChapterThreeWithNoArtifactDealsNothing(t *testing.T) {
	g, _ := mshA1ArmorWarsTable(t, 0)
	me, opp := g.Seats[0], g.Seats[1]
	hobA9NextChapter(t, g)
	hobA9NextChapter(t, g)
	life := opp.Life
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d -> %d, want no change with no artifact", life, opp.Life)
	}
}

// --- Avengers Assemble! -------------------------------------------------------------

func TestMshA1AvengersAssembleGivesYourHeroesPlusTwoPlusTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Avengers Assemble!", "Enchantment", mshA1AssembleOracle, false)
	hero := b12Creature(g, me.ID, "My Hero", "Creature — Human Hero", 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirHero := b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	if got := hobA4Power(t, g, hero); got != 4 {
		t.Errorf("my Hero has power %d, want 4", got)
	}
	if got := hobA4Power(t, g, bear); got != 2 {
		t.Errorf("my Bear has power %d, want 2", got)
	}
	if got := hobA4Power(t, g, theirHero); got != 2 {
		t.Errorf("their Hero has power %d, want 2", got)
	}
}

func TestMshA1AvengersAssembleHasFlash(t *testing.T) {
	spec, _ := Lookup(mshA1AssembleOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestMshA1AvengersAssembleDrawsAtEndStepAfterAHeroAttacked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Avengers Assemble!", "Enchantment", mshA1AssembleOracle, false)
	hero := b12Creature(g, me.ID, "My Hero", "Creature — Human Hero", 2, 2)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, hero)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestMshA1AvengersAssembleDrawsAtEndStepAfterAHeroEntered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Avengers Assemble!", "Enchantment", mshA1AssembleOracle, false)
	hand := me.Hand.Size()
	b12PlayFromHand(t, g, "Test Hero", "Creature — Human Hero", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	// b12PlayFromHand adds the Hero to the hand and casts it, so the net
	// change is the trigger's one card.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1", got)
	}
}

func TestMshA1AvengersAssembleDrawsNothingWithoutAHeroThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Avengers Assemble!", "Enchantment", mshA1AssembleOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Old Hero", "Creature — Human Hero", 2, 2) // here already, and stays home
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, bear)
	b12PlayFromHand(t, g, "Test Bear", "Creature — Bear", "", game.CastSpellParams{})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand changed by %d, want 0: no Hero attacked or entered", got)
	}
}

// --- Avengers Tower -------------------------------------------------------------------

func TestMshA1AvengersTowerManaIsLockedToHeroSpellsAndHeroAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tower := pushCatalogPermanent(g, me.ID, "Avengers Tower", "Land", mshA1TowerOracle, false)

	// The first ability is plain {C}.
	if err := g.ActivateManaAbility(me.ID, tower, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility {C}: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("pool %v, want one {C}", got)
	}
	if len(me.ManaPool.Units()[0].Restrictions) != 0 {
		t.Error("the colourless half carries a restriction")
	}
	me.ManaPool = game.ManaPool{}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(tower) })

	if err := g.ActivateManaAbility(me.ID, tower, 1, game.ManaAbilityParams{Colors: []string{"W"}}); err != nil {
		t.Fatalf("ActivateManaAbility any colour: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "W" {
		t.Fatalf("pool %v, want one {W}", got)
	}
	cost, _ := game.ParseCost("{W}")
	hero := game.Card{Name: "Test Hero", TypeLine: "Creature — Human Hero", ManaCost: "{W}"}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{W}"}
	rite := game.Card{Name: "Test Rite", TypeLine: "Sorcery", ManaCost: "{W}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(hero)) {
		t.Error("the mana was refused for a Hero spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) {
		t.Error("the mana paid for a non-Hero creature spell: stronger than printed")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rite)) {
		t.Error("the mana paid for a sorcery: stronger than printed")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(hero)) {
		t.Error("the mana was refused for an ability of a Hero source")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana paid for an ability of a non-Hero source: stronger than printed")
	}
}

func TestMshA1AvengersTowerLooksAtThreeAndTakesAHero(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tower := pushCatalogPermanent(g, me.ID, "Avengers Tower", "Land", mshA1TowerOracle, false)
	advanceToMain(t, g)
	hero := uuid.New()
	ids := seedLibrary(me, "A", "B")
	me.Library.PushTop(game.Card{InstanceID: hero, Name: "Test Hero", TypeLine: "Creature — Human Hero", Owner: me.ID, Controller: me.ID})
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tower, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMax != 1 {
		t.Fatalf("no take-one prompt: %+v", pick)
	}
	if hasID(pick.ChooseCards, ids[0]) || hasID(pick.ChooseCards, ids[1]) || !hasID(pick.ChooseCards, hero) {
		t.Errorf("offered %v: only the Hero card is a legal pick", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{hero}); err != nil {
		t.Fatalf("take the Hero: %v", err)
	}
	if !me.Hand.Contains(hero) {
		t.Fatal("the Hero card is not in hand")
	}
	order := putInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("the rest are not ordered on the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[1], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := libraryBottomIDs(me, 2); !sameIDs(got, []uuid.UUID{ids[1], ids[0]}) {
		t.Errorf("bottom = %v, want the player's order", got)
	}
}

// --- Baxter Building --------------------------------------------------------------------

func TestMshA1BaxterBuildingMakesFourManaOfAnyColoursForFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	baxter := pushCatalogPermanent(g, me.ID, "Baxter Building", "Land", mshA1BaxterOracle, false)
	apaMana(me, "C", "C", "C")
	err := g.ActivateManaAbility(me.ID, baxter, 1, game.ManaAbilityParams{Colors: []string{"W", "U", "B", "R"}})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("three mana: err = %v, want ErrInsufficientMana", err)
	}
	if b16Tapped(t, g, baxter) {
		t.Fatal("a refused activation tapped the Building")
	}
	apaMana(me, "C")
	if err := g.ActivateManaAbility(me.ID, baxter, 1, game.ManaAbilityParams{Colors: []string{"W", "U", "B", "R"}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := poolColors(me); len(got) != 4 {
		t.Errorf("pool %v, want four mana", got)
	}
	if !b16Tapped(t, g, baxter) {
		t.Error("the Building did not tap")
	}
}

func TestMshA1BaxterBuildingDrawsOnlyWithAToughnessFourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	baxter := pushCatalogPermanent(g, me.ID, "Baxter Building", "Land", mshA1BaxterOracle, false)
	small := acPermanent(g, me.ID, "Fast Bear", "Creature — Bear", 5, 3)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C")

	acRefused(t, g, me.ID, baxter, 0, game.ActivateAbilityParams{})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(small, game.CounterPlusOne, 1) })
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, baxter, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn once a creature has toughness 4", hand, me.Hand.Size())
	}
}

// --- Black Panther, Vanguard --------------------------------------------------------------

// mshA1HeroEnters plays a Hero from the hand with Black Panther on the table
// and lets the stack settle up to the mode question.
func mshA1HeroEnters(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	b12PlayFromHand(t, g, "Test Hero", "Creature — Human Hero", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	return latestChoiceOfKindFor(g, game.PendingChoiceModePick, g.Seats[0].ID)
}

func TestMshA1BlackPantherMakesASoldierWhenANontokenHeroEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Black Panther, Vanguard", "Legendary Creature — Human Warrior Hero", mshA1BlackPantherOracle, false)
	c := mshA1HeroEnters(t, g)
	if c == nil {
		t.Fatal("a Hero entering asked for no mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countBattlefieldNamed(g, me.ID, "Soldier") != 1 {
		t.Error("the first mode made no 1/1 Soldier")
	}
}

func TestMshA1BlackPantherPumpsYourCreaturesIncludingTheHeroThatEntered(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Black Panther, Vanguard", "Legendary Creature — Human Warrior Hero", mshA1BlackPantherOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c := mshA1HeroEnters(t, g)
	if c == nil {
		t.Fatal("a Hero entering asked for no mode")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := hobA4Power(t, g, bear); got != 3 {
		t.Errorf("my Bear has power %d, want 3", got)
	}
	if got := hobA4Power(t, g, theirs); got != 2 {
		t.Errorf("their Bear has power %d, want 2", got)
	}
	if countBattlefieldNamed(g, me.ID, "Soldier") != 0 {
		t.Error("the second mode made a Soldier")
	}
}

func TestMshA1BlackPantherIgnoresTokensNonHeroesAndHimself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Black Panther, Vanguard", "Legendary Creature — Human Warrior Hero", mshA1BlackPantherOracle, false)

	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, game.Card{Name: "Hero", TypeLine: "Token Creature — Human Hero", Power: 1, Toughness: 1}, 1)
	})
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("a Hero token triggered Black Panther")
	}

	b12PlayFromHand(t, g, "Test Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("a non-Hero creature triggered Black Panther")
	}
}

func TestMshA1BlackPantherDoesNotTriggerOnHimself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castHoldCreature(t, g, "Black Panther, Vanguard", "Legendary Creature — Human Warrior Hero", mshA1BlackPantherOracle, 4, 4)
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("Black Panther entering asked for a mode: the trigger says another Hero")
	}
}

func TestMshA1BlackPantherIgnoresAnOpponentsHero(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Black Panther, Vanguard", "Legendary Creature — Human Warrior Hero", mshA1BlackPantherOracle, false)
	theirHero := b12Creature(g, opp.ID, "Their Hero", "Creature — Human Hero", 2, 2)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: theirHero, Actor: opp.ID})
	})
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) != nil {
		t.Error("an opponent's Hero triggered Black Panther")
	}
}
