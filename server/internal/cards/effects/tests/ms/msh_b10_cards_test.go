package testsms

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func TestMshB10TonyStarkTakesAnArtifactFromTheTopFourAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tony := effectstest.PushBattlefieldCardWithTimestamp(g, mshB10TonyCard(me.ID))
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Deep Card", TypeLine: "Creature — Bear"},
		game.Card{Name: "Top Rock", TypeLine: "Artifact"},
		game.Card{Name: "Top Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Top Gem", TypeLine: "Artifact"},
		game.Card{Name: "Top Wolf", TypeLine: "Creature — Wolf"},
	)
	// seedSearchLibrary puts the FIRST card on top.
	deep, rock, bear, gem, wolf := ids[0], ids[1], ids[2], ids[3], ids[4]
	_ = deep

	if err := effectstest.HobB4Activate(t, g, me.ID, tony, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt for the look")
	}
	if len(c.ChooseCards) != 2 {
		t.Errorf("%d cards offered, want only the two artifacts among the top four", len(c.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, gem)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(gem) {
		t.Error("the artifact is not in hand")
	}
	for _, id := range []uuid.UUID{rock, bear, wolf} {
		if !me.Library.Contains(id) || me.Hand.Contains(id) {
			t.Errorf("card %s should be on the bottom of the library", id)
		}
	}
	if !effectstest.B16Tapped(t, g, tony) {
		t.Error("the ability did not tap Tony Stark")
	}
}

func TestMshB10TonyStarkMayTakeNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tony := effectstest.PushBattlefieldCardWithTimestamp(g, mshB10TonyCard(me.ID))
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Fifth", TypeLine: "Creature — Bear"},
		game.Card{Name: "A", TypeLine: "Artifact"},
		game.Card{Name: "B", TypeLine: "Creature — Bear"},
		game.Card{Name: "C", TypeLine: "Creature — Bear"},
		game.Card{Name: "D", TypeLine: "Creature — Bear"},
	)
	hand := me.Hand.Size()
	if err := effectstest.HobB4Activate(t, g, me.ID, tony, 0, "{C}{C}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no card taken", hand, me.Hand.Size())
	}
}

func TestMshB10TonyStarkTransformsForSixAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tony := effectstest.PushBattlefieldCardWithTimestamp(g, mshB10TonyCard(me.ID))
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := effectstest.HobB4Activate(t, g, me.ID, tony, 1, "{C}{C}{C}{C}{U}{R}", nil); err == nil {
		t.Fatal("the transform was activated outside a main phase")
	}
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	tony = effectstest.PushBattlefieldCardWithTimestamp(g, mshB10TonyCard(me.ID))
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.HobB4Activate(t, g, me.ID, tony, 1, "{C}{C}{C}{C}{U}{R}", nil); err != nil {
		t.Fatalf("transform: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, tony)
	if !ok || c.Name != "The Invincible Iron Man" {
		t.Fatalf("after the transform: %+v", c)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, tony, "flying") || !effectstest.EffectiveAbilitiesContain(t, g, tony, "haste") {
		t.Error("The Invincible Iron Man lacks flying or haste")
	}
}

func mshB10IronManCombat(t *testing.T, handCards ...game.Card) (g *game.Game, me *game.Player, ironMan uuid.UUID, hand []uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	ironMan = mshB10IronMan(g, me.ID)
	for _, c := range handCards {
		id := uuid.New()
		c.InstanceID, c.Owner, c.Controller = id, me.ID, me.ID
		g.WithWriteLock(func() { me.Hand.PushTop(c) })
		hand = append(hand, id)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, ironMan, hand
}

func TestMshB10IronManPutsAnEquipmentFromHandAndAttachesIt(t *testing.T) {
	g, me, ironMan, hand := mshB10IronManCombat(t,
		game.Card{Name: "Test Blade", TypeLine: "Artifact — Equipment"},
		game.Card{Name: "Test Bear", TypeLine: "Creature — Bear"},
	)
	blade := hand[0]
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no prompt to put an artifact card from hand")
	}
	if len(c.ChooseCards) != 1 {
		t.Errorf("%d cards offered, want only the artifact", len(c.ChooseCards))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, blade)
	effectstest.PassPriorityAroundTable(t, g)

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
	effectstest.AnswerChooseCards(t, g, me.ID, hand[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hand[0]) {
		t.Fatal("the artifact did not enter")
	}
	if n := len(g.AttachmentsOf(ironMan)); n != 0 {
		t.Errorf("%d attachments on The Invincible Iron Man, want none", n)
	}
}

func TestMshB10IronManMayPutNothing(t *testing.T) {
	g, me, _, hand := mshB10IronManCombat(t, game.Card{Name: "Test Rock", TypeLine: "Artifact"})
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hand[0]) || !me.Hand.Contains(hand[0]) {
		t.Error("declining still put the artifact onto the battlefield")
	}
}

// --- Training Regimen ----------------------------------------------------------

func TestMshB10RegimenGivesTrampleToCreaturesWithPlusOneCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Training Regimen", "Enchantment", mshB10RegimenOracle, 0, 0)
	plain := effectstest.B12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	counted := effectstest.B12Creature(g, me.ID, "Counted Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{counted, theirs} {
		g.WithWriteLock(func() { _, _ = g.AddCounterMustSettleNowForEffect(id, game.CounterPlusOne, 1) })
	}
	if effectstest.EffectiveAbilitiesContain(t, g, plain, "trample") {
		t.Error("a creature with no counters has trample")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, counted, "trample") {
		t.Error("a creature with a +1/+1 counter lacks trample")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "trample") {
		t.Error("the opponent's countered creature has trample")
	}
}

func TestMshB10RegimenPutsACounterOnATargetAtYourBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Training Regimen", "Enchantment", mshB10RegimenOracle, 0, 0)
	first := effectstest.B12Creature(g, me.ID, "First Bear", "Creature — Bear", 2, 2)
	second := effectstest.B12Creature(g, me.ID, "Second Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.MshB2PickIfAsked(t, g, me.ID, second)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, second); got != 1 {
		t.Errorf("chosen creature has %d counters, want 1", got)
	}
	if got := effectstest.PlusOnes(g, first); got != 0 {
		t.Errorf("unchosen creature has %d counters, want 0", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, second, "trample") {
		t.Error("the creature that got the counter has no trample")
	}
}

// --- Truck Toss -----------------------------------------------------------------

func TestMshB10TruckTossCostsTwoLessWithAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, mshB10TruckTossOracle, "Instant", "{2}{R}{R}"); got != 4 {
		t.Errorf("without a Vehicle: %d, want 4", got)
	}
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Test Truck", TypeLine: "Artifact — Vehicle"})
	if got := effectstest.SelfPricedMV(t, g, me, mshB10TruckTossOracle, "Instant", "{2}{R}{R}"); got != 2 {
		t.Errorf("with a Vehicle: %d, want 2", got)
	}
}

func TestMshB10TruckTossDealsFourToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Truck Toss", "Instant", mshB10TruckTossOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("life %d -> %d, want -4", before, opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Truck Toss", "Instant", mshB10TruckTossOracle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("4 damage killed a 5/5")
	}
}

// --- Undercover Skrull ----------------------------------------------------------

func TestMshB10SkrullGrowsAndIsEveryTypeWithTwoCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skrull := effectstest.B12Push(g, me.ID, "Undercover Skrull", "Creature — Skrull Shapeshifter Villain", mshB10SkrullOracle, 1, 1)
	effectstest.KillCreature(t, g, me.ID, effectstest.B12Creature(g, me.ID, "Dead Bear", "Creature — Bear", 2, 2))
	effectstest.KillCreature(t, g, me.ID, effectstest.B12Permanent(g, me.ID, "Dead Rock", "Artifact"))
	if p, tough := effectstest.EffectivePTOf(t, g, skrull); p != 1 || tough != 1 {
		t.Fatalf("Skrull is %d/%d with one creature card in the graveyard, want 1/1", p, tough)
	}
	if effectstest.RingBFCard(t, g, skrull).HasSubtype("Wizard") {
		t.Fatal("Skrull is every creature type with one creature card in the graveyard")
	}
	effectstest.KillCreature(t, g, me.ID, effectstest.B12Creature(g, me.ID, "Dead Wolf", "Creature — Wolf", 2, 2))
	if p, tough := effectstest.EffectivePTOf(t, g, skrull); p != 3 || tough != 3 {
		t.Errorf("Skrull is %d/%d with two creature cards in the graveyard, want 3/3", p, tough)
	}
	if !effectstest.RingBFCard(t, g, skrull).HasSubtype("Wizard") {
		t.Error("Skrull is not every creature type with two creature cards in the graveyard")
	}
}

func TestMshB10SkrullTapsForManaOfAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	skrull := effectstest.B12Push(g, me.ID, "Undercover Skrull", "Creature — Skrull Shapeshifter Villain", mshB10SkrullOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, skrull, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool = %v, want one red mana", got)
	}
	if !effectstest.B16Tapped(t, g, skrull) {
		t.Error("the mana ability did not tap the Skrull")
	}
}

// --- Vibranium Energy Daggers ---------------------------------------------------

func TestMshB10DaggersGiveTwoTwoAndAreIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	daggers := effectstest.B12Push(g, me.ID, "Vibranium Energy Daggers", "Artifact — Equipment", mshB10DaggersOracle, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.HobB4Activate(t, g, me.ID, daggers, 0, "{C}{C}{C}", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.LgKnowLibrary(me)
	robot := effectstest.PushGraveyardPermanent(me, "Buried Robot", "Artifact Creature — Robot", "{3}")
	bear := effectstest.PushGraveyardPermanent(me, "Buried Bear", "Creature — Bear", "{2}")
	tooBig := effectstest.PushGraveyardPermanent(me, "Buried Titan", "Artifact Creature — Golem", "{6}")

	effectstest.CastXSpell(t, g, "Vision Quest", "Sorcery", mshB10VisionQuestOrcl, "{X}{U}{R}", 4, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
	effectstest.AnswerSearchByID(t, g, me.ID, robot)

	if !g.Battlefield.Contains(robot) {
		t.Fatal("the artifact creature did not enter")
	}
	if got := effectstest.PlusOnes(g, robot); got != 4 {
		t.Errorf("+1/+1 counters = %d, want X = 4", got)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, robot, "haste") {
		t.Error("X = 4 did not give haste")
	}
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(tooBig) {
		t.Error("a non-artifact creature or one above X entered")
	}
	if effectstest.SearchEventCount(g) != 0 || effectstest.LibraryKnownTo(me, me.ID) != 1 {
		t.Error("a graveyard-only search searched or shuffled the library")
	}
}

func TestMshB10VisionQuestBelowFourGivesNoHasteAndFindsInTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Library Robot", TypeLine: "Artifact Creature — Robot", ManaCost: "{2}", Power: 2, Toughness: 2},
	)
	robot := ids[1]

	effectstest.CastXSpell(t, g, "Vision Quest", "Sorcery", mshB10VisionQuestOrcl, "{X}{U}{R}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, robot)

	if !g.Battlefield.Contains(robot) {
		t.Fatal("the library artifact creature did not enter")
	}
	if got := effectstest.PlusOnes(g, robot); got != 2 {
		t.Errorf("+1/+1 counters = %d, want X = 2", got)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, robot, "haste") {
		t.Error("X = 2 gave haste")
	}
	if effectstest.SearchEventCount(g) == 0 {
		t.Error("searching the library fired no search event")
	}
}

// --- Visions of Villainy --------------------------------------------------------

func TestMshB10VisionsCostsOneLessWithAVillain(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, mshB10VisionsOracle, "Instant", "{2}{B}"); got != 3 {
		t.Errorf("without a Villain: %d, want 3", got)
	}
	effectstest.B12Creature(g, me.ID, "Test Villain", "Creature — Human Villain", 2, 2)
	if got := effectstest.SelfPricedMV(t, g, me, mshB10VisionsOracle, "Instant", "{2}{B}"); got != 2 {
		t.Errorf("with a Villain: %d, want 2", got)
	}
}

func TestMshB10VisionsDrawsTwoAndLosesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		game.Card{Name: "A", TypeLine: "Creature — Bear"},
		game.Card{Name: "B", TypeLine: "Creature — Bear"},
		game.Card{Name: "C", TypeLine: "Creature — Bear"},
	)
	hand, life := me.Hand.Size(), me.Life
	effectstest.CastCatalogSpell(t, g, "Visions of Villainy", "Instant", mshB10VisionsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2", life, me.Life)
	}
}

// --- Volcanic Villain -----------------------------------------------------------

func TestMshB10VolcanicVillainHasHasteAndPowersUpForThreeTheTurnItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	villain := effectstest.CastPowerUpCreature(t, g, me, "Volcanic Villain", "Creature — Elemental Villain", "{2}{R}", "{C}{C}{R}", mshB10VolcanicOracle, 3)
	if !effectstest.EffectiveAbilitiesContain(t, g, villain, "haste") {
		t.Error("Volcanic Villain lacks haste")
	}
	if got := effectstest.ChargedPowerUp(t, g, me.ID, villain); got != "{3}" {
		t.Fatalf("the turn it entered: %q, want {3}", got)
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, villain, 0, "{C}{C}{C}", nil); err != nil {
		t.Fatalf("power up: %v", err)
	}
	effectstest.ResolveTopOfStack(t, g)
	if got := effectstest.PlusOnes(g, villain); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

func TestMshB10VolcanicVillainPaysFullPriceTheNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.PowerUpMain(t, g)
	me := g.Seats[0]
	villain := effectstest.CastPowerUpCreature(t, g, me, "Volcanic Villain", "Creature — Elemental Villain", "{2}{R}", "{C}{C}{R}", mshB10VolcanicOracle, 3)
	effectstest.PowerUpPassTo(t, g, 0)
	if got := effectstest.ChargedPowerUp(t, g, me.ID, villain); got != "{5}{R}" {
		t.Errorf("the next turn: %q, want {5}{R}", got)
	}
}

// --- Wakandan Royal Guard -------------------------------------------------------

func mshB10CastGuardOn(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	guard := effectstest.CastCatalogSpell(t, g, "Wakandan Royal Guard", "Creature — Human Soldier Hero", mshB10RoyalGuardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MshB2PickIfAsked(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		target := effectstest.B12Creature(g, me.ID, tc.name, tc.typeLine, 2, 2)
		effectstest.B12Creature(g, me.ID, "Spare", "Creature — Bear", 2, 2)
		mshB10CastGuardOn(t, g, target)
		if got := effectstest.PlusOnes(g, target); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.name, got, tc.want)
		}
	}
}

func TestMshB10RoyalGuardTargetingItselfGetsOnlyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Spare", "Creature — Bear", 2, 2)
	guard := effectstest.CastCatalogSpell(t, g, "Wakandan Royal Guard", "Creature — Human Soldier Hero", mshB10RoyalGuardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MshB2PickIfAsked(t, g, me.ID, guard)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, guard); got != 1 {
		t.Errorf("the Guard targeting itself got %d counters, want 1 (it is not another Hero)", got)
	}
}
