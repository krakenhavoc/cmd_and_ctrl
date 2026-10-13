package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// villainous_cards_test.go — #3140: the seven cards that use a
// villainous choice and are fully expressible with what the engine has.

const (
	villainThisIsHowItEnds = "53467cf4-5db2-4b9c-af2b-03b044c558cd"
	villainSycorax         = "78ed3e01-6d22-429a-a9db-baff38d7bc03"
	villainDavros          = "25c0bbc8-b0c3-4166-967f-e8b096487a6c"
	villainEggman          = "fb68c1ae-a129-4c17-81fa-be5da2c32ec5"
	villainGreatPlan       = "cb9abc41-07f8-4a06-b7b9-9dd6be526e42"
)

// villainClearHand empties a player's hand and library of the harness's
// starting cards, so the counts a test makes are its own.
func villainEmptyHand(p *game.Player) { p.Hand.Cards = nil }

// ---- Damocles Base, Sword of Kang -----------------------------------------

func TestDamoclesBaseMakesTheDamagedPlayerFaceAVillainousChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	base := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Damocles Base, Sword of Kang", OracleID: villainDamocles,
		TypeLine: "Legendary Artifact Creature — Vehicle", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	villainFillLibrary(g, me, 6)

	attackWith(t, g, opp.ID, base)
	passPriorityAroundTable(t, g)

	c := villainPrompt(t, g, opp.ID)
	lifeBefore, handBefore := opp.Life, villainHandSize(me)
	villainAnswer(t, g, c, 1)
	if opp.Life != lifeBefore-2 || villainHandSize(me) != handBefore+2 {
		t.Errorf("lose 2 / draw 2: life %d -> %d, hand %d -> %d", lifeBefore, opp.Life, handBefore, villainHandSize(me))
	}
	// Crew 3 is printed and present.
	spec, _ := Lookup(villainDamocles)
	crew := false
	for _, a := range spec.Activated {
		crew = crew || a.Cost.Crew == 3
	}
	if !crew {
		t.Error("Crew 3 is missing")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("completeness = %v", spec.Completeness)
	}
}

func TestDamoclesBaseWithTheValeyardAsksTheDamagedPlayerTwice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	base := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Damocles Base, Sword of Kang", OracleID: villainDamocles,
		TypeLine: "Legendary Artifact Creature — Vehicle", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	b43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	villainFillLibrary(g, me, 6)

	attackWith(t, g, opp.ID, base)
	passPriorityAroundTable(t, g)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)
	lifeBefore := opp.Life
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)
	if opp.Life != lifeBefore-2 {
		t.Errorf("the additional time did not run: life %d -> %d", lifeBefore, opp.Life)
	}
}

// ---- The Dalek Emperor ---------------------------------------------------------

func TestTheDalekEmperorAsksEachOpponentAtTheBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Dalek Emperor", OracleID: villainDalekEmperor,
		TypeLine: "Legendary Artifact Creature — Dalek", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	bears := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats[1:] {
		bears[p.ID] = pushVanillaCreature(g, p.ID, "Bear", 2, 2)
	}

	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	// Opponent 1 sacrifices; opponent 2 lets me have a Dalek; opponent 3
	// sacrifices too. One at a time, in turn order.
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[1].ID), 0)
	answerSacrifice(t, g, g.Seats[1].ID, bears[g.Seats[1].ID])
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[2].ID), 1)
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[3].ID), 0)
	answerSacrifice(t, g, g.Seats[3].ID, bears[g.Seats[3].ID])

	if findBattlefieldCardForTest(g, bears[g.Seats[1].ID]) != nil || findBattlefieldCardForTest(g, bears[g.Seats[3].ID]) != nil {
		t.Error("a sacrificed Bear is still on the battlefield")
	}
	if findBattlefieldCardForTest(g, bears[g.Seats[2].ID]) == nil {
		t.Error("the opponent who chose the token lost their Bear")
	}
	if n := battlefieldCountNamed(g, me.ID, "Dalek"); n != 1 {
		t.Errorf("%d Dalek tokens, want 1", n)
	}
}

func TestTheDalekEmperorGivesOtherDaleksYouControlHasteAndHasAffinity(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	emperor := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Dalek Emperor", OracleID: villainDalekEmperor,
		TypeLine: "Legendary Artifact Creature — Dalek", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3,
		Owner: opp.ID, Controller: opp.ID,
	})
	if !effectiveAbilitiesContain(t, g, mine, "haste") {
		t.Error("another Dalek I control has no haste")
	}
	if effectiveAbilitiesContain(t, g, emperor, "haste") {
		t.Error("the Emperor itself was given haste; it says 'other'")
	}
	if effectiveAbilitiesContain(t, g, theirs, "haste") {
		t.Error("an opponent's Dalek was given haste")
	}
	// Affinity for Daleks: {5}{B}{R} less one for each of my two Daleks
	// (the Emperor and the token).
	if got := selfPricedMV(t, g, me, villainDalekEmperor, "Legendary Artifact Creature — Dalek", "{5}{B}{R}"); got != 5 {
		t.Errorf("the Emperor costs %d with two Daleks, want 5", got)
	}
}

// ---- This Is How It Ends ------------------------------------------------------

func villainCastEnds(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "This Is How It Ends", "Instant", villainThisIsHowItEnds, cardTarget(target))
	passPriorityAroundTable(t, g)
}

func TestThisIsHowItEndsShufflesTheCreatureThenItsOwnerFacesTheChoice(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	libBefore := opp.Library.Size()
	lifeBefore := opp.Life

	villainCastEnds(t, g, bear)

	if findBattlefieldCardForTest(g, bear) != nil || opp.Library.Size() != libBefore+1 {
		t.Fatalf("the Bear was not shuffled into its owner's library (library %d -> %d)", libBefore, opp.Library.Size())
	}
	c := villainPrompt(t, g, opp.ID)
	if c.PickOptions[0].Label != "Lose 5 life" || c.PickOptions[0].LifeCost != 5 {
		t.Errorf("first option = %+v", c.PickOptions[0])
	}
	// With no other creature the second option does nothing and is free.
	if c.PickOptions[1].Harm != 0 {
		t.Errorf("shuffling a creature you do not have is priced as a cost: %+v", c.PickOptions[1])
	}
	villainAnswer(t, g, c, 0)
	if opp.Life != lifeBefore-5 {
		t.Errorf("lost %d life, want 5", lifeBefore-opp.Life)
	}
}

func TestThisIsHowItEndsSecondOptionShufflesAnotherCreatureTheyOwn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	wolf := pushVanillaCreature(g, opp.ID, "Wolf", 3, 3)
	other := pushVanillaCreature(g, g.Seats[2].ID, "Elk", 2, 2)
	libBefore := opp.Library.Size()
	lifeBefore := opp.Life

	villainCastEnds(t, g, bear)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)

	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("no creature pick: %+v", g.PendingChoices)
	}
	for _, id := range pick.ChooseCards {
		if id == bear || id == other {
			t.Errorf("offered %s: the shuffled Bear and another player's creature are not 'another creature they own'", id)
		}
	}
	answerChooseCards(t, g, opp.ID, wolf)
	if findBattlefieldCardForTest(g, wolf) != nil {
		t.Error("the Wolf is still on the battlefield")
	}
	if findBattlefieldCardForTest(g, other) == nil {
		t.Error("another player's creature was taken")
	}
	if opp.Library.Size() != libBefore+2 || opp.Life != lifeBefore {
		t.Errorf("library %d -> %d, life %d -> %d", libBefore, opp.Library.Size(), lifeBefore, opp.Life)
	}
}

// The OWNER faces it, not the creature's controller and not the caster.
func TestThisIsHowItEndsIsFacedByTheOwnerOfAStolenCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	stolen := pushDiesCreatureForTest(g, owner.ID, "Stolen", "", "Creature — Test", 2, 2)
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, stolen).Controller = me.ID
	})

	villainCastEnds(t, g, stolen)

	villainPrompt(t, g, owner.ID)
}

// ---- Sycorax Commander ---------------------------------------------------------

func villainCastSycorax(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castCatalogSpell(t, g, "Sycorax Commander", "Creature — Alien Soldier", villainSycorax, nil)
	passPriorityAroundTable(t, g)
	return id
}

func TestSycoraxCommanderFirstOptionDiscardsTheHandAndDrawsOneFewer(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	villainEmptyHand(opp)
	for i := 0; i < 3; i++ {
		handCardForTest(opp, "Card", "Sorcery", "")
	}
	villainFillLibrary(g, opp, 6)
	for _, p := range g.Seats[2:] {
		villainEmptyHand(p)
	}

	villainCastSycorax(t, g)

	c := villainPrompt(t, g, opp.ID)
	villainAnswer(t, g, c, 0)
	if got := villainHandSize(opp); got != 2 {
		t.Errorf("hand = %d, want 2 (discarded 3, drew 3 minus one)", got)
	}
}

func TestSycoraxCommanderSecondOptionDealsDamageEqualToTheHand(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	villainEmptyHand(opp)
	for i := 0; i < 3; i++ {
		handCardForTest(opp, "Card", "Sorcery", "")
	}
	lifeBefore := opp.Life
	for _, p := range g.Seats[2:] {
		villainEmptyHand(p)
	}

	villainCastSycorax(t, g)
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)

	if opp.Life != lifeBefore-3 {
		t.Errorf("took %d damage, want 3", lifeBefore-opp.Life)
	}
	if villainHandSize(opp) != 3 {
		t.Errorf("damage option touched the hand: %d", villainHandSize(opp))
	}
}

func TestSycoraxCommanderWithAnEmptyHandBothOptionsDoNothing(t *testing.T) {
	g := newCatalogGame(t)
	for _, p := range g.Seats[1:] {
		villainEmptyHand(p)
	}
	opp := g.Seats[1]
	lifeBefore := opp.Life

	villainCastSycorax(t, g)
	c := villainPrompt(t, g, opp.ID)
	if len(c.PickOptions) != 2 || c.PickOptions[0].Harm != 0 || c.PickOptions[1].Harm != 0 {
		t.Errorf("an empty hand makes both options free: %+v", c.PickOptions)
	}
	villainAnswer(t, g, c, 0)
	if opp.Life != lifeBefore || villainHandSize(opp) != 0 {
		t.Errorf("an empty hand took %d damage / drew %d", lifeBefore-opp.Life, villainHandSize(opp))
	}
}

func TestSycoraxCommanderIsFirstStrikeAndHaste(t *testing.T) {
	spec, ok := Lookup(villainSycorax)
	if !ok || len(spec.PrintedKeywords) != 2 || spec.PrintedKeywords[0] != "first strike" || spec.PrintedKeywords[1] != "haste" {
		t.Errorf("keywords = %v", spec.PrintedKeywords)
	}
}

// ---- Dr. Eggman ----------------------------------------------------------------

func villainEggmanAtEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dr. Eggman", OracleID: villainEggman,
		TypeLine: "Legendary Creature — Human Scientist", Power: 3, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
}

func TestDrEggmanDrawsThenEachOpponentFacesTheChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	handBefore := villainHandSize(me)
	for _, p := range g.Seats[2:] {
		villainEmptyHand(p)
	}
	for i := 0; i < 2; i++ {
		handCardForTest(opp, "Card", "Sorcery", "")
	}
	oppHand := villainHandSize(opp)

	villainEggmanAtEndStep(t, g)

	if villainHandSize(me) != handBefore+1 {
		t.Errorf("Eggman's controller drew %d, want 1", villainHandSize(me)-handBefore)
	}
	// Opponent 1 discards.
	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 0)
	discardFromHand(t, g, opp.ID)
	if villainHandSize(opp) != oppHand-1 {
		t.Errorf("hand %d -> %d, want one discarded", oppHand, villainHandSize(opp))
	}
	// Opponent 2 lets me put a Construct onto the battlefield.
	construct := handCardForTest(me, "Servo Bot", "Artifact Creature — Construct", "")
	handCardForTest(me, "Bear", "Creature — Bear", "")
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[2].ID), 1)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the controller is not offered the put: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != construct {
		t.Errorf("offered %v, want only the Construct", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, construct)
	if findBattlefieldCardForTest(g, construct) == nil {
		t.Error("the Construct did not enter the battlefield")
	}
	// Opponent 3 with an empty hand: the discard does nothing.
	villainAnswer(t, g, villainPrompt(t, g, g.Seats[3].ID), 0)
	villainNoPrompt(t, g)
}

func TestDrEggmanDoesNotTellABotWhatIsInTheControllersHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats[1:] {
		villainEmptyHand(p)
	}
	handCardForTest(g.Seats[1], "Card", "Sorcery", "")
	villainEmptyHand(me)
	villainEggmanAtEndStep(t, g)
	without := villainPrompt(t, g, g.Seats[1].ID).PickOptions[1].Harm

	g2 := newCatalogGame(t)
	for _, p := range g2.Seats[1:] {
		villainEmptyHand(p)
	}
	handCardForTest(g2.Seats[1], "Card", "Sorcery", "")
	handCardForTest(g2.Seats[0], "Servo Bot", "Artifact Creature — Construct", "")
	villainEggmanAtEndStep(t, g2)
	with := villainPrompt(t, g2, g2.Seats[1].ID).PickOptions[1].Harm

	if without != with {
		t.Errorf("the hint on the controller's option reads their hidden hand: %d vs %d", without, with)
	}
}

// ---- Davros, Dalek Creator --------------------------------------------------

func villainDavrosAtEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Davros, Dalek Creator", OracleID: villainDavros,
		TypeLine: "Legendary Artifact Creature — Alien Scientist", Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	advanceToEndStepOf(t, g, 0)
}

func villainLose(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, -n); err != nil {
			panic(err)
		}
	})
}

func TestDavrosCreatesADalekAndAsksOnlyOpponentsWhoLostThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me, hurt, scratched := g.Seats[0], g.Seats[1], g.Seats[2]
	villainFillLibrary(g, me, 4)
	villainLose(g, hurt, 3)
	villainLose(g, scratched, 2)
	villainDavrosAtEndStep(t, g)
	passPriorityAroundTable(t, g)

	if n := battlefieldCountNamed(g, me.ID, "Dalek"); n != 1 {
		t.Fatalf("%d Dalek tokens, want 1", n)
	}
	c := villainPrompt(t, g, hurt.ID)
	handBefore := villainHandSize(me)
	villainAnswer(t, g, c, 0)
	if villainHandSize(me) != handBefore+1 {
		t.Errorf("'you draw a card' gave the controller %d cards", villainHandSize(me)-handBefore)
	}
	villainNoPrompt(t, g)
}

func TestDavrosDiscardOptionIsTheFacingPlayersOwnDiscard(t *testing.T) {
	g := newCatalogGame(t)
	hurt := g.Seats[1]
	villainLose(g, hurt, 4)
	handBefore := villainHandSize(hurt)
	villainDavrosAtEndStep(t, g)
	passPriorityAroundTable(t, g)

	villainAnswer(t, g, villainPrompt(t, g, hurt.ID), 1)
	discardFromHand(t, g, hurt.ID)
	if villainHandSize(hurt) != handBefore-1 {
		t.Errorf("hand %d -> %d", handBefore, villainHandSize(hurt))
	}
}

func TestDavrosDoesNothingWhenNoOpponentLostThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	villainLose(g, g.Seats[1], 2)
	villainDavrosAtEndStep(t, g)
	passPriorityAroundTable(t, g)

	if n := battlefieldCountNamed(g, me.ID, "Dalek"); n != 0 {
		t.Errorf("%d Dalek tokens with nobody at 3 life lost", n)
	}
	villainNoPrompt(t, g)
}

// ---- Great Intelligence's Plan ---------------------------------------------------

func villainCastPlan(t *testing.T, g *game.Game, target *game.Player) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Great Intelligence's Plan", TypeLine: "Sorcery", OracleID: villainGreatPlan,
		Owner: active.ID, Controller: active.ID,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestGreatIntelligencesPlanDrawsThreeThenTheTargetFacesTheChoice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	villainFillLibrary(g, me, 6)
	villainEmptyHand(me)
	for i := 0; i < 5; i++ {
		handCardForTest(opp, "Card", "Sorcery", "")
	}
	oppBefore := villainHandSize(opp)

	villainCastPlan(t, g, opp)

	if villainHandSize(me) != 3 {
		t.Errorf("the controller holds %d cards, want 3 drawn", villainHandSize(me))
	}
	c := villainPrompt(t, g, opp.ID)
	villainAnswer(t, g, c, 0)
	discardFromHand(t, g, opp.ID)
	if got := villainHandSize(opp); got != oppBefore-3 {
		t.Errorf("hand %d -> %d, want three discarded", oppBefore, got)
	}
}

func TestGreatIntelligencesPlanSecondOptionLetsTheControllerCastForFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	villainFillLibrary(g, me, 6)
	villainEmptyHand(me)
	villainCastPlan(t, g, opp)

	villainAnswer(t, g, villainPrompt(t, g, opp.ID), 1)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the controller was not offered the free cast: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 {
		t.Errorf("the free cast is a 'may': floor = %d", pick.ChooseMin)
	}
	for _, id := range pick.ChooseCards {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			t.Error("a land was offered as a spell to cast")
		}
	}
}

func TestGreatIntelligencesPlanRequiresAnOpponentTarget(t *testing.T) {
	spec, ok := Lookup(villainGreatPlan)
	if !ok || spec.Targets == nil || spec.Targets.Label != "target opponent" {
		t.Errorf("targets = %+v", spec.Targets)
	}
}
