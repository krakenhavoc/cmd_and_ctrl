package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ---- Damocles Base, Sword of Kang -----------------------------------------

func TestDamoclesBaseMakesTheDamagedPlayerFaceAVillainousChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	base := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Damocles Base, Sword of Kang", OracleID: villainDamocles,
		TypeLine: "Legendary Artifact Creature — Vehicle", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	effectstest.VillainFillLibrary(g, me, 6)

	effectstest.AttackWith(t, g, opp.ID, base)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.VillainPrompt(t, g, opp.ID)
	lifeBefore, handBefore := opp.Life, villainHandSize(me)
	effectstest.VillainAnswer(t, g, c, 1)
	if opp.Life != lifeBefore-2 || villainHandSize(me) != handBefore+2 {
		t.Errorf("lose 2 / draw 2: life %d -> %d, hand %d -> %d", lifeBefore, opp.Life, handBefore, villainHandSize(me))
	}
	// Crew 3 is printed and present.
	spec, _ := effects.Lookup(villainDamocles)
	crew := false
	for _, a := range spec.Activated {
		crew = crew || a.Cost.Crew == 3
	}
	if !crew {
		t.Error("Crew 3 is missing")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("completeness = %v", spec.Completeness)
	}
}

func TestDamoclesBaseWithTheValeyardAsksTheDamagedPlayerTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	base := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Damocles Base, Sword of Kang", OracleID: villainDamocles,
		TypeLine: "Legendary Artifact Creature — Vehicle", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID,
	})
	effectstest.B43Catalog(g, me.ID, "The Valeyard", "Legendary Creature — Time Lord Noble", villainValeyard, 4, 5)
	effectstest.VillainFillLibrary(g, me, 6)

	effectstest.AttackWith(t, g, opp.ID, base)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)
	lifeBefore := opp.Life
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)
	if opp.Life != lifeBefore-2 {
		t.Errorf("the additional time did not run: life %d -> %d", lifeBefore, opp.Life)
	}
}

// ---- The Dalek Emperor ---------------------------------------------------------

func TestTheDalekEmperorAsksEachOpponentAtTheBeginningOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Dalek Emperor", OracleID: villainDalekEmperor,
		TypeLine: "Legendary Artifact Creature — Dalek", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	bears := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats[1:] {
		bears[p.ID] = effectstest.PushVanillaCreature(g, p.ID, "Bear", 2, 2)
	}

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)

	// Opponent 1 sacrifices; opponent 2 lets me have a Dalek; opponent 3
	// sacrifices too. One at a time, in turn order.
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[1].ID), 0)
	effectstest.AnswerSacrifice(t, g, g.Seats[1].ID, bears[g.Seats[1].ID])
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[2].ID), 1)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[3].ID), 0)
	effectstest.AnswerSacrifice(t, g, g.Seats[3].ID, bears[g.Seats[3].ID])

	if effectstest.FindBattlefieldCardForTest(g, bears[g.Seats[1].ID]) != nil || effectstest.FindBattlefieldCardForTest(g, bears[g.Seats[3].ID]) != nil {
		t.Error("a sacrificed Bear is still on the battlefield")
	}
	if effectstest.FindBattlefieldCardForTest(g, bears[g.Seats[2].ID]) == nil {
		t.Error("the opponent who chose the token lost their Bear")
	}
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Dalek"); n != 1 {
		t.Errorf("%d Dalek tokens, want 1", n)
	}
}

func TestTheDalekEmperorGivesOtherDaleksYouControlHasteAndHasAffinity(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	emperor := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Dalek Emperor", OracleID: villainDalekEmperor,
		TypeLine: "Legendary Artifact Creature — Dalek", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dalek", TypeLine: "Token Artifact Creature — Dalek", Power: 3, Toughness: 3,
		Owner: opp.ID, Controller: opp.ID,
	})
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "haste") {
		t.Error("another Dalek I control has no haste")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, emperor, "haste") {
		t.Error("the Emperor itself was given haste; it says 'other'")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "haste") {
		t.Error("an opponent's Dalek was given haste")
	}
	// Affinity for Daleks: {5}{B}{R} less one for each of my two Daleks
	// (the Emperor and the token).
	if got := effectstest.SelfPricedMV(t, g, me, villainDalekEmperor, "Legendary Artifact Creature — Dalek", "{5}{B}{R}"); got != 5 {
		t.Errorf("the Emperor costs %d with two Daleks, want 5", got)
	}
}

func TestThisIsHowItEndsShufflesTheCreatureThenItsOwnerFacesTheChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	libBefore := opp.Library.Size()
	lifeBefore := opp.Life

	effectstest.VillainCastEnds(t, g, bear)

	if effectstest.FindBattlefieldCardForTest(g, bear) != nil || opp.Library.Size() != libBefore+1 {
		t.Fatalf("the Bear was not shuffled into its owner's library (library %d -> %d)", libBefore, opp.Library.Size())
	}
	c := effectstest.VillainPrompt(t, g, opp.ID)
	if c.PickOptions[0].Label != "Lose 5 life" || c.PickOptions[0].LifeCost != 5 {
		t.Errorf("first option = %+v", c.PickOptions[0])
	}
	// With no other creature the second option does nothing and is free.
	if c.PickOptions[1].Harm != 0 {
		t.Errorf("shuffling a creature you do not have is priced as a cost: %+v", c.PickOptions[1])
	}
	effectstest.VillainAnswer(t, g, c, 0)
	if opp.Life != lifeBefore-5 {
		t.Errorf("lost %d life, want 5", lifeBefore-opp.Life)
	}
}

func TestThisIsHowItEndsSecondOptionShufflesAnotherCreatureTheyOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 2)
	wolf := effectstest.PushVanillaCreature(g, opp.ID, "Wolf", 3, 3)
	other := effectstest.PushVanillaCreature(g, g.Seats[2].ID, "Elk", 2, 2)
	libBefore := opp.Library.Size()
	lifeBefore := opp.Life

	effectstest.VillainCastEnds(t, g, bear)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)

	pick := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("no creature pick: %+v", g.PendingChoices)
	}
	for _, id := range pick.ChooseCards {
		if id == bear || id == other {
			t.Errorf("offered %s: the shuffled Bear and another player's creature are not 'another creature they own'", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, wolf)
	if effectstest.FindBattlefieldCardForTest(g, wolf) != nil {
		t.Error("the Wolf is still on the battlefield")
	}
	if effectstest.FindBattlefieldCardForTest(g, other) == nil {
		t.Error("another player's creature was taken")
	}
	if opp.Library.Size() != libBefore+2 || opp.Life != lifeBefore {
		t.Errorf("library %d -> %d, life %d -> %d", libBefore, opp.Library.Size(), lifeBefore, opp.Life)
	}
}

// The OWNER faces it, not the creature's controller and not the caster.
func TestThisIsHowItEndsIsFacedByTheOwnerOfAStolenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	stolen := effectstest.PushDiesCreatureForTest(g, owner.ID, "Stolen", "", "Creature — Test", 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, stolen).Controller = me.ID
	})

	effectstest.VillainCastEnds(t, g, stolen)

	effectstest.VillainPrompt(t, g, owner.ID)
}

// ---- Sycorax Commander ---------------------------------------------------------

func villainCastSycorax(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Sycorax Commander", "Creature — Alien Soldier", effectstest.VillainSycorax, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestSycoraxCommanderFirstOptionDiscardsTheHandAndDrawsOneFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.VillainEmptyHand(opp)
	for i := 0; i < 3; i++ {
		effectstest.HandCardForTest(opp, "Card", "Sorcery", "")
	}
	effectstest.VillainFillLibrary(g, opp, 6)
	for _, p := range g.Seats[2:] {
		effectstest.VillainEmptyHand(p)
	}

	villainCastSycorax(t, g)

	c := effectstest.VillainPrompt(t, g, opp.ID)
	effectstest.VillainAnswer(t, g, c, 0)
	if got := villainHandSize(opp); got != 2 {
		t.Errorf("hand = %d, want 2 (discarded 3, drew 3 minus one)", got)
	}
}

func TestSycoraxCommanderSecondOptionDealsDamageEqualToTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.VillainEmptyHand(opp)
	for i := 0; i < 3; i++ {
		effectstest.HandCardForTest(opp, "Card", "Sorcery", "")
	}
	lifeBefore := opp.Life
	for _, p := range g.Seats[2:] {
		effectstest.VillainEmptyHand(p)
	}

	villainCastSycorax(t, g)
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)

	if opp.Life != lifeBefore-3 {
		t.Errorf("took %d damage, want 3", lifeBefore-opp.Life)
	}
	if villainHandSize(opp) != 3 {
		t.Errorf("damage option touched the hand: %d", villainHandSize(opp))
	}
}

func TestSycoraxCommanderWithAnEmptyHandBothOptionsDoNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	for _, p := range g.Seats[1:] {
		effectstest.VillainEmptyHand(p)
	}
	opp := g.Seats[1]
	lifeBefore := opp.Life

	villainCastSycorax(t, g)
	c := effectstest.VillainPrompt(t, g, opp.ID)
	if len(c.PickOptions) != 2 || c.PickOptions[0].Harm != 0 || c.PickOptions[1].Harm != 0 {
		t.Errorf("an empty hand makes both options free: %+v", c.PickOptions)
	}
	effectstest.VillainAnswer(t, g, c, 0)
	if opp.Life != lifeBefore || villainHandSize(opp) != 0 {
		t.Errorf("an empty hand took %d damage / drew %d", lifeBefore-opp.Life, villainHandSize(opp))
	}
}

func TestSycoraxCommanderIsFirstStrikeAndHaste(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.VillainSycorax)
	if !ok || len(spec.PrintedKeywords) != 2 || spec.PrintedKeywords[0] != "first strike" || spec.PrintedKeywords[1] != "haste" {
		t.Errorf("keywords = %v", spec.PrintedKeywords)
	}
}

// ---- Dr. Eggman ----------------------------------------------------------------

func villainEggmanAtEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dr. Eggman", OracleID: effectstest.VillainEggman,
		TypeLine: "Legendary Creature — Human Scientist", Power: 3, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestDrEggmanDrawsThenEachOpponentFacesTheChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	handBefore := villainHandSize(me)
	for _, p := range g.Seats[2:] {
		effectstest.VillainEmptyHand(p)
	}
	for i := 0; i < 2; i++ {
		effectstest.HandCardForTest(opp, "Card", "Sorcery", "")
	}
	oppHand := villainHandSize(opp)

	villainEggmanAtEndStep(t, g)

	if villainHandSize(me) != handBefore+1 {
		t.Errorf("Eggman's controller drew %d, want 1", villainHandSize(me)-handBefore)
	}
	// Opponent 1 discards.
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 0)
	effectstest.DiscardFromHand(t, g, opp.ID)
	if villainHandSize(opp) != oppHand-1 {
		t.Errorf("hand %d -> %d, want one discarded", oppHand, villainHandSize(opp))
	}
	// Opponent 2 lets me put a Construct onto the battlefield.
	construct := effectstest.HandCardForTest(me, "Servo Bot", "Artifact Creature — Construct", "")
	effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[2].ID), 1)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("the controller is not offered the put: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != construct {
		t.Errorf("offered %v, want only the Construct", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, construct)
	if effectstest.FindBattlefieldCardForTest(g, construct) == nil {
		t.Error("the Construct did not enter the battlefield")
	}
	// Opponent 3 with an empty hand: the discard does nothing.
	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, g.Seats[3].ID), 0)
	villainNoPrompt(t, g)
}

func TestDrEggmanDoesNotTellABotWhatIsInTheControllersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats[1:] {
		effectstest.VillainEmptyHand(p)
	}
	effectstest.HandCardForTest(g.Seats[1], "Card", "Sorcery", "")
	effectstest.VillainEmptyHand(me)
	villainEggmanAtEndStep(t, g)
	without := effectstest.VillainPrompt(t, g, g.Seats[1].ID).PickOptions[1].Harm

	g2 := effectstest.NewCatalogGame(t)
	for _, p := range g2.Seats[1:] {
		effectstest.VillainEmptyHand(p)
	}
	effectstest.HandCardForTest(g2.Seats[1], "Card", "Sorcery", "")
	effectstest.HandCardForTest(g2.Seats[0], "Servo Bot", "Artifact Creature — Construct", "")
	villainEggmanAtEndStep(t, g2)
	with := effectstest.VillainPrompt(t, g2, g2.Seats[1].ID).PickOptions[1].Harm

	if without != with {
		t.Errorf("the hint on the controller's option reads their hidden hand: %d vs %d", without, with)
	}
}

// ---- Davros, Dalek Creator --------------------------------------------------

func villainDavrosAtEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	me := g.Seats[0]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Davros, Dalek Creator", OracleID: effectstest.VillainDavros,
		TypeLine: "Legendary Artifact Creature — Alien Scientist", Power: 3, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	effectstest.AdvanceToEndStepOf(t, g, 0)
}

func villainLose(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, -n); err != nil {
			panic(err)
		}
	})
}

func TestDavrosCreatesADalekAndAsksOnlyOpponentsWhoLostThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, hurt, scratched := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.VillainFillLibrary(g, me, 4)
	villainLose(g, hurt, 3)
	villainLose(g, scratched, 2)
	villainDavrosAtEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Dalek"); n != 1 {
		t.Fatalf("%d Dalek tokens, want 1", n)
	}
	c := effectstest.VillainPrompt(t, g, hurt.ID)
	handBefore := villainHandSize(me)
	effectstest.VillainAnswer(t, g, c, 0)
	if villainHandSize(me) != handBefore+1 {
		t.Errorf("'you draw a card' gave the controller %d cards", villainHandSize(me)-handBefore)
	}
	villainNoPrompt(t, g)
}

func TestDavrosDiscardOptionIsTheFacingPlayersOwnDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	hurt := g.Seats[1]
	villainLose(g, hurt, 4)
	handBefore := villainHandSize(hurt)
	villainDavrosAtEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, hurt.ID), 1)
	effectstest.DiscardFromHand(t, g, hurt.ID)
	if villainHandSize(hurt) != handBefore-1 {
		t.Errorf("hand %d -> %d", handBefore, villainHandSize(hurt))
	}
}

func TestDavrosDoesNothingWhenNoOpponentLostThreeLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	villainLose(g, g.Seats[1], 2)
	villainDavrosAtEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Dalek"); n != 0 {
		t.Errorf("%d Dalek tokens with nobody at 3 life lost", n)
	}
	villainNoPrompt(t, g)
}

func TestGreatIntelligencesPlanDrawsThreeThenTheTargetFacesTheChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.VillainFillLibrary(g, me, 6)
	effectstest.VillainEmptyHand(me)
	for i := 0; i < 5; i++ {
		effectstest.HandCardForTest(opp, "Card", "Sorcery", "")
	}
	oppBefore := villainHandSize(opp)

	effectstest.VillainCastPlan(t, g, opp)

	if villainHandSize(me) != 3 {
		t.Errorf("the controller holds %d cards, want 3 drawn", villainHandSize(me))
	}
	c := effectstest.VillainPrompt(t, g, opp.ID)
	effectstest.VillainAnswer(t, g, c, 0)
	effectstest.DiscardFromHand(t, g, opp.ID)
	if got := villainHandSize(opp); got != oppBefore-3 {
		t.Errorf("hand %d -> %d, want three discarded", oppBefore, got)
	}
}

func TestGreatIntelligencesPlanSecondOptionLetsTheControllerCastForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.VillainFillLibrary(g, me, 6)
	effectstest.VillainEmptyHand(me)
	effectstest.VillainCastPlan(t, g, opp)

	effectstest.VillainAnswer(t, g, effectstest.VillainPrompt(t, g, opp.ID), 1)
	// #4250: the free cast is offered inside the spell's resolution, and
	// declining it (the "may") is always an answer.
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast)
	if len(offer.DiscardOptions) == 0 {
		t.Fatal("the controller was offered no card to cast")
	}
	for _, id := range offer.DiscardOptions {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
			t.Error("a land was offered as a spell to cast")
		}
	}
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if g.ResolutionPaused() {
		t.Error("the decline did not let the spell finish resolving")
	}
}

func TestGreatIntelligencesPlanRequiresAnOpponentTarget(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.VillainGreatPlan)
	if !ok || spec.Targets == nil || spec.Targets.Label != "target opponent" {
		t.Errorf("targets = %+v", spec.Targets)
	}
}
