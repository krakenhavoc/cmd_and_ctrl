package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_b10_cards_test.go — Marvel Super Heroes, half B, tenth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshB10TonyOracle       = "4a46ea59-6c5d-4bed-9951-b5590f9a1bdb"
	mshB10RegimenOracle    = "f44bde8b-43e9-4719-ac60-b8f0cee8df9b"
	mshB10TruckTossOracle  = "7d765055-da62-43f7-9580-98e9d164753b"
	mshB10SkrullOracle     = "33b1379f-b329-42cd-92ed-4811d44ed8dd"
	mshB10DaggersOracle    = "be0b86fc-8450-41d4-9836-6de11cecb857"
	mshB10VisionQuestOrcl  = "240410a0-3288-4336-b1ad-34936e64ccf2"
	mshB10VisionsOracle    = "3b46800b-711b-4cbb-b296-26115b4f6839"
	mshB10VolcanicOracle   = "a257426d-70c9-4d58-bd8b-97ec38f6e133"
	mshB10RoyalGuardOracle = "4ae15bd1-e0ea-4c1a-a311-dae421381ebc"
)

func TestMshB10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mshB10TonyOracle:        "Tony Stark",
		mshB10TonyOracle + "#1": "The Invincible Iron Man",
		mshB10RegimenOracle:     "Training Regimen",
		mshB10TruckTossOracle:   "Truck Toss",
		mshB10SkrullOracle:      "Undercover Skrull",
		mshB10DaggersOracle:     "Vibranium Energy Daggers",
		mshB10VisionQuestOrcl:   "Vision Quest",
		mshB10VisionsOracle:     "Visions of Villainy",
		mshB10VolcanicOracle:    "Volcanic Villain",
		mshB10RoyalGuardOracle:  "Wakandan Royal Guard",
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

// --- Tony Stark // The Invincible Iron Man ------------------------------------

func mshB10TonyCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   mshB10TonyOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Tony Stark", TypeLine: "Legendary Creature — Human Artificer Hero", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 1, Toughness: 3},
			{Name: "The Invincible Iron Man", TypeLine: "Legendary Artifact Creature — Human Hero", Colors: []string{"U", "R"}, Power: 4, Toughness: 5},
		},
	}
	c.SetFace(0)
	return c
}

func mshB10IronMan(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := mshB10TonyCard(owner)
	c.SetFace(1)
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestMshB10TonyStarkTakesAnArtifactFromTheTopFourAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tony := pushBattlefieldCardWithTimestamp(g, mshB10TonyCard(me.ID))
	advanceToMain(t, g)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Deep Card", TypeLine: "Creature — Bear"},
		game.Card{Name: "Top Rock", TypeLine: "Artifact"},
		game.Card{Name: "Top Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Top Gem", TypeLine: "Artifact"},
		game.Card{Name: "Top Wolf", TypeLine: "Creature — Wolf"},
	)
	// seedSearchLibrary puts the FIRST card on top.
	deep, rock, bear, gem, wolf := ids[0], ids[1], ids[2], ids[3], ids[4]
	_ = deep

	if err := hobB4Activate(t, g, me.ID, tony, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt for the look")
	}
	if len(c.ChooseCards) != 2 {
		t.Errorf("%d cards offered, want only the two artifacts among the top four", len(c.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, gem)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(gem) {
		t.Error("the artifact is not in hand")
	}
	for _, id := range []uuid.UUID{rock, bear, wolf} {
		if !me.Library.Contains(id) || me.Hand.Contains(id) {
			t.Errorf("card %s should be on the bottom of the library", id)
		}
	}
	if !b16Tapped(t, g, tony) {
		t.Error("the ability did not tap Tony Stark")
	}
}

func TestMshB10TonyStarkMayTakeNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tony := pushBattlefieldCardWithTimestamp(g, mshB10TonyCard(me.ID))
	advanceToMain(t, g)
	seedSearchLibrary(me,
		game.Card{Name: "Fifth", TypeLine: "Creature — Bear"},
		game.Card{Name: "A", TypeLine: "Artifact"},
		game.Card{Name: "B", TypeLine: "Creature — Bear"},
		game.Card{Name: "C", TypeLine: "Creature — Bear"},
		game.Card{Name: "D", TypeLine: "Creature — Bear"},
	)
	hand := me.Hand.Size()
	if err := hobB4Activate(t, g, me.ID, tony, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no card taken", hand, me.Hand.Size())
	}
}

func TestMshB10TonyStarkTransformsForSixAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tony := pushBattlefieldCardWithTimestamp(g, mshB10TonyCard(me.ID))
	advanceTo(t, g, game.StepBeginCombat)
	if err := hobB4Activate(t, g, me.ID, tony, 1, "{C}{C}{C}{C}{U}{R}", nil); err == nil {
		t.Fatal("the transform was activated outside a main phase")
	}
	g = newCatalogGame(t)
	me = g.Seats[0]
	tony = pushBattlefieldCardWithTimestamp(g, mshB10TonyCard(me.ID))
	advanceToMain(t, g)
	if err := hobB4Activate(t, g, me.ID, tony, 1, "{C}{C}{C}{C}{U}{R}", nil); err != nil {
		t.Fatalf("transform: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, tony)
	if !ok || c.Name != "The Invincible Iron Man" {
		t.Fatalf("after the transform: %+v", c)
	}
	if !effectiveAbilitiesContain(t, g, tony, "flying") || !effectiveAbilitiesContain(t, g, tony, "haste") {
		t.Error("The Invincible Iron Man lacks flying or haste")
	}
}

func mshB10IronManCombat(t *testing.T, handCards ...game.Card) (g *game.Game, me *game.Player, ironMan uuid.UUID, hand []uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	ironMan = mshB10IronMan(g, me.ID)
	for _, c := range handCards {
		id := uuid.New()
		c.InstanceID, c.Owner, c.Controller = id, me.ID, me.ID
		g.WithWriteLock(func() { me.Hand.PushTop(c) })
		hand = append(hand, id)
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	return g, me, ironMan, hand
}

func TestMshB10IronManPutsAnEquipmentFromHandAndAttachesIt(t *testing.T) {
	g, me, ironMan, hand := mshB10IronManCombat(t,
		game.Card{Name: "Test Blade", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Test Bear", TypeLine: "Creature — Bear"},
	)
	blade := hand[0]
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt to put an artifact card from hand")
	}
	if len(c.ChooseCards) != 1 {
		t.Errorf("%d cards offered, want only the artifact", len(c.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, blade)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(blade) {
		t.Fatal("the Equipment did not enter")
	}
	found := false
	for _, id := range g.AttachmentsOf(ironMan) {
		if id == blade {
			found = true
		}
	}
	if !found {
		t.Error("the Equipment was not attached to The Invincible Iron Man")
	}
}

func TestMshB10IronManDoesNotAttachANonEquipmentArtifact(t *testing.T) {
	g, me, ironMan, hand := mshB10IronManCombat(t, game.Card{Name: "Test Rock", TypeLine: "Artifact"})
	answerChooseCards(t, g, me.ID, hand[0])
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hand[0]) {
		t.Fatal("the artifact did not enter")
	}
	if n := len(g.AttachmentsOf(ironMan)); n != 0 {
		t.Errorf("%d attachments on The Invincible Iron Man, want none", n)
	}
}

func TestMshB10IronManMayPutNothing(t *testing.T) {
	g, me, _, hand := mshB10IronManCombat(t, game.Card{Name: "Test Rock", TypeLine: "Artifact"})
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hand[0]) || !me.Hand.Contains(hand[0]) {
		t.Error("declining still put the artifact onto the battlefield")
	}
}

// --- Training Regimen ----------------------------------------------------------

func TestMshB10RegimenGivesTrampleToCreaturesWithPlusOneCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Training Regimen", "Enchantment", mshB10RegimenOracle, 0, 0)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	counted := b12Creature(g, me.ID, "Counted Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{counted, theirs} {
		g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(id, game.CounterPlusOne, 1) })
	}
	if effectiveAbilitiesContain(t, g, plain, "trample") {
		t.Error("a creature with no counters has trample")
	}
	if !effectiveAbilitiesContain(t, g, counted, "trample") {
		t.Error("a creature with a +1/+1 counter lacks trample")
	}
	if effectiveAbilitiesContain(t, g, theirs, "trample") {
		t.Error("the opponent's countered creature has trample")
	}
}

func TestMshB10RegimenPutsACounterOnATargetAtYourBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Training Regimen", "Enchantment", mshB10RegimenOracle, 0, 0)
	first := b12Creature(g, me.ID, "First Bear", "Creature — Bear", 2, 2)
	second := b12Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	mshB2PickIfAsked(t, g, me.ID, second)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, second); got != 1 {
		t.Errorf("chosen creature has %d counters, want 1", got)
	}
	if got := plusOnes(g, first); got != 0 {
		t.Errorf("unchosen creature has %d counters, want 0", got)
	}
	if !effectiveAbilitiesContain(t, g, second, "trample") {
		t.Error("the creature that got the counter has no trample")
	}
}

// --- Truck Toss -----------------------------------------------------------------

func TestMshB10TruckTossCostsTwoLessWithAVehicle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, mshB10TruckTossOracle, "Instant", "{2}{R}{R}"); got != 4 {
		t.Errorf("without a Vehicle: %d, want 4", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Test Truck", TypeLine: "Artifact — Vehicle"})
	if got := selfPricedMV(t, g, me, mshB10TruckTossOracle, "Instant", "{2}{R}{R}"); got != 2 {
		t.Errorf("with a Vehicle: %d, want 2", got)
	}
}

func TestMshB10TruckTossDealsFourToAnyTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := b12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	before := opp.Life
	castCatalogSpell(t, g, "Truck Toss", "Instant", mshB10TruckTossOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("life %d -> %d, want -4", before, opp.Life)
	}
	castCatalogSpell(t, g, "Truck Toss", "Instant", mshB10TruckTossOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("4 damage killed a 5/5")
	}
}

// --- Undercover Skrull ----------------------------------------------------------

func TestMshB10SkrullGrowsAndIsEveryTypeWithTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skrull := b12Push(g, me.ID, "Undercover Skrull", "Creature — Skrull Shapeshifter Villain", mshB10SkrullOracle, 1, 1)
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Dead Bear", "Creature — Bear", 2, 2))
	killCreature(t, g, me.ID, b12Permanent(g, me.ID, "Dead Rock", "Artifact"))
	if p, tough := effectivePTOf(t, g, skrull); p != 1 || tough != 1 {
		t.Fatalf("Skrull is %d/%d with one creature card in the graveyard, want 1/1", p, tough)
	}
	if ringBFCard(t, g, skrull).HasSubtype("Wizard") {
		t.Fatal("Skrull is every creature type with one creature card in the graveyard")
	}
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Dead Wolf", "Creature — Wolf", 2, 2))
	if p, tough := effectivePTOf(t, g, skrull); p != 3 || tough != 3 {
		t.Errorf("Skrull is %d/%d with two creature cards in the graveyard, want 3/3", p, tough)
	}
	if !ringBFCard(t, g, skrull).HasSubtype("Wizard") {
		t.Error("Skrull is not every creature type with two creature cards in the graveyard")
	}
}

func TestMshB10SkrullTapsForManaOfAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	skrull := b12Push(g, me.ID, "Undercover Skrull", "Creature — Skrull Shapeshifter Villain", mshB10SkrullOracle, 1, 1)
	advanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, skrull, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool = %v, want one red mana", got)
	}
	if !b16Tapped(t, g, skrull) {
		t.Error("the mana ability did not tap the Skrull")
	}
}

// --- Vibranium Energy Daggers ---------------------------------------------------

func TestMshB10DaggersGiveTwoTwoAndAreIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	daggers := b12Push(g, me.ID, "Vibranium Energy Daggers", "Artifact — Equipment", mshB10DaggersOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := hobB4Activate(t, g, me.ID, daggers, 0, "{C}{C}{C}", cardRefs(bear)); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("equipped Bear is %d/%d, want 4/4", p, tough)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(daggers) })
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(daggers) {
		t.Error("the Daggers were destroyed")
	}
}

// --- Vision Quest ---------------------------------------------------------------

func TestMshB10VisionQuestPutsAGraveyardArtifactCreatureInWithXCountersAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	lgKnowLibrary(me)
	robot := pushGraveyardPermanent(me, "Buried Robot", "Artifact Creature — Robot", "{3}")
	bear := pushGraveyardPermanent(me, "Buried Bear", "Creature — Bear", "{2}")
	tooBig := pushGraveyardPermanent(me, "Buried Titan", "Artifact Creature — Golem", "{6}")

	castXSpell(t, g, "Vision Quest", "Sorcery", mshB10VisionQuestOrcl, "{X}{U}{R}", 4, nil)
	passPriorityAroundTable(t, g)
	answerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	answerSearchByID(t, g, me.ID, robot)

	if !g.Battlefield.Contains(robot) {
		t.Fatal("the artifact creature did not enter")
	}
	if got := plusOnes(g, robot); got != 4 {
		t.Errorf("+1/+1 counters = %d, want X = 4", got)
	}
	if !effectiveAbilitiesContain(t, g, robot, "haste") {
		t.Error("X = 4 did not give haste")
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(tooBig) {
		t.Error("a non-artifact creature or one above X entered")
	}
	if searchEventCount(g) != 0 || libraryKnownTo(me, me.ID) != 1 {
		t.Error("a graveyard-only search searched or shuffled the library")
	}
}

func TestMshB10VisionQuestBelowFourGivesNoHasteAndFindsInTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Library Robot", TypeLine: "Artifact Creature — Robot", ManaCost: "{2}", Power: 2, Toughness: 2},
	)
	robot := ids[1]

	castXSpell(t, g, "Vision Quest", "Sorcery", mshB10VisionQuestOrcl, "{X}{U}{R}", 2, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, robot)

	if !g.Battlefield.Contains(robot) {
		t.Fatal("the library artifact creature did not enter")
	}
	if got := plusOnes(g, robot); got != 2 {
		t.Errorf("+1/+1 counters = %d, want X = 2", got)
	}
	if effectiveAbilitiesContain(t, g, robot, "haste") {
		t.Error("X = 2 gave haste")
	}
	if searchEventCount(g) == 0 {
		t.Error("searching the library fired no search event")
	}
}

// --- Visions of Villainy --------------------------------------------------------

func TestMshB10VisionsCostsOneLessWithAVillain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, mshB10VisionsOracle, "Instant", "{2}{B}"); got != 3 {
		t.Errorf("without a Villain: %d, want 3", got)
	}
	b12Creature(g, me.ID, "Test Villain", "Creature — Human Villain", 2, 2)
	if got := selfPricedMV(t, g, me, mshB10VisionsOracle, "Instant", "{2}{B}"); got != 2 {
		t.Errorf("with a Villain: %d, want 2", got)
	}
}

func TestMshB10VisionsDrawsTwoAndLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Creature — Bear"},
		game.Card{Name: "B", TypeLine: "Creature — Bear"},
		game.Card{Name: "C", TypeLine: "Creature — Bear"},
	)
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Visions of Villainy", "Instant", mshB10VisionsOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
}

// --- Volcanic Villain -----------------------------------------------------------

func TestMshB10VolcanicVillainHasHasteAndPowersUpForThreeTheTurnItEnters(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	villain := castPowerUpCreature(t, g, me, "Volcanic Villain", "Creature — Elemental Villain", "{2}{R}", "{C}{C}{R}", mshB10VolcanicOracle, 3)
	if !effectiveAbilitiesContain(t, g, villain, "haste") {
		t.Error("Volcanic Villain lacks haste")
	}
	if got := chargedPowerUp(t, g, me.ID, villain); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3}", got)
	}
	if err := hobB4Activate(t, g, me.ID, villain, 0, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, villain); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

func TestMshB10VolcanicVillainPaysFullPriceTheNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	villain := castPowerUpCreature(t, g, me, "Volcanic Villain", "Creature — Elemental Villain", "{2}{R}", "{C}{C}{R}", mshB10VolcanicOracle, 3)
	powerUpPassTo(t, g, 0)
	if got := chargedPowerUp(t, g, me.ID, villain); got != "{5}{R}" {
		t.Errorf("the next turn: %q, want {5}{R}", got)
	}
}

// --- Wakandan Royal Guard -------------------------------------------------------

func mshB10CastGuardOn(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	guard := castCatalogSpell(t, g, "Wakandan Royal Guard", "Creature — Human Soldier Hero", mshB10RoyalGuardOracle, nil)
	passPriorityAroundTable(t, g)
	mshB2PickIfAsked(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	return guard
}

func TestMshB10RoyalGuardPutsOneCounterOnANonHeroAndTwoOnAnotherHero(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		want           int
	}{
		{"Plain Bear", "Creature — Bear", 1},
		{"Other Hero", "Creature — Human Hero", 2},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		target := b12Creature(g, me.ID, tc.name, tc.typeLine, 2, 2)
		b12Creature(g, me.ID, "Spare", "Creature — Bear", 2, 2)
		mshB10CastGuardOn(t, g, target)
		if got := plusOnes(g, target); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.name, got, tc.want)
		}
	}
}

func TestMshB10RoyalGuardTargetingItselfGetsOnlyOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Creature(g, me.ID, "Spare", "Creature — Bear", 2, 2)
	guard := castCatalogSpell(t, g, "Wakandan Royal Guard", "Creature — Human Soldier Hero", mshB10RoyalGuardOracle, nil)
	passPriorityAroundTable(t, g)
	mshB2PickIfAsked(t, g, me.ID, guard)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, guard); got != 1 {
		t.Errorf("the Guard targeting itself got %d counters, want 1 (it is not another Hero)", got)
	}
}
