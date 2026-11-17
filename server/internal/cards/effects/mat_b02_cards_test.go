package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mat_b02_cards_test.go — March of the Machine: The Aftermath, batch 2
// (#5476). Every test plays the real catalog entry through the engine.

const (
	matCosmicRebirthOracle     = "add60aad-2532-4f59-bfb9-f0168c2adda6"
	matCampusRenovationOracle  = "d34a3b58-a905-4e90-b35c-9f5c21149b04"
	matHarnessedSnubhornOracle = "e1dd9dca-cb5f-46b4-b16d-a61ae651d653"
	matDeathRattleOniOracle    = "bea40c67-73ca-4b5e-b902-5de07f21585c"
	matFeastOracle             = "7e4e1d13-db1a-48e5-9b5f-f3e71d10c23e"
	matUndercityUpheavalOracle = "c19f4a43-85bb-4070-a237-171597c94ec3"
	matTolarianContemptOracle  = "831623c9-8279-439e-a42b-ac47cacdb058"
	matTranquilFrillbackOracle = "e87a4461-f3d1-443f-ad4c-dba91a993809"
	matNahirisResolveOracle    = "12bd2051-297b-4bab-ac64-6eb26a42e3e2"
)

func TestMatB02AllNineAreFullWithNoCaveats(t *testing.T) {
	for _, o := range []string{matCosmicRebirthOracle, matCampusRenovationOracle, matHarnessedSnubhornOracle,
		matDeathRattleOniOracle, matFeastOracle, matUndercityUpheavalOracle, matTolarianContemptOracle,
		matTranquilFrillbackOracle, matNahirisResolveOracle} {
		spec, ok := Lookup(o)
		if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want a Full catalog entry with no caveats, got %+v", o, spec)
		}
	}
}

// --- Cosmic Rebirth ------------------------------------------------------

func TestMatB02CosmicRebirthPutsASmallPermanentOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	rock := wthGraveyardCard(me, "Old Rock", "Artifact")
	life := me.Life
	castCatalogSpell(t, g, "Cosmic Rebirth", "Instant", matCosmicRebirthOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "Old Rock"); !ok {
		t.Error("the mana value 0 card was not put onto the battlefield")
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestMatB02CosmicRebirthDecliningPutsItIntoYourHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	rock := wthGraveyardCard(me, "Old Rock", "Artifact")
	life := me.Life
	castCatalogSpell(t, g, "Cosmic Rebirth", "Instant", matCosmicRebirthOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, me, rock); zone != "hand" {
		t.Errorf("the card is in %s, want the hand", zone)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestMatB02CosmicRebirthAnExpensiveCardGoesStraightToHand(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	big := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}", Owner: me.ID, Controller: me.ID})
	life := me.Life
	castCatalogSpell(t, g, "Cosmic Rebirth", "Instant", matCosmicRebirthOracle, cardTarget(big))
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("asked whether to put a mana value 5 card onto the battlefield")
	}
	if zone, _ := cardWhere(g, me, big); zone != "hand" {
		t.Errorf("the card is in %s, want the hand", zone)
	}
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestMatB02CosmicRebirthRefusesAnInstantCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bolt := wthGraveyardCard(me, "Bolt", "Instant")
	if err := castCatalogSpellErr(t, g, "Cosmic Rebirth", "Instant", matCosmicRebirthOracle, cardTarget(bolt)); err == nil {
		t.Fatal("targeted a non-permanent card")
	}
}

// --- Campus Renovation ---------------------------------------------------

func TestMatB02CampusRenovationReturnsThenExilesTwoToPlay(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	rock := wthGraveyardCard(me, "Old Rock", "Artifact")
	seedLibrary(me, "Top One", "Top Two", "Deep")
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Campus Renovation", "Sorcery", matCampusRenovationOracle, cardTarget(rock))
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "Old Rock"); !ok {
		t.Error("the artifact card did not return to the battlefield")
	}
	if got := lib - me.Library.Size(); got != 2 {
		t.Errorf("exiled %d cards from the library, want 2", got)
	}
	if len(g.Exile.Cards) < 2 {
		t.Errorf("exile holds %d cards, want the two exiled", len(g.Exile.Cards))
	}
}

func TestMatB02CampusRenovationWorksWithNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	seedLibrary(me, "Top One", "Top Two", "Deep")
	lib := me.Library.Size()
	castCatalogSpell(t, g, "Campus Renovation", "Sorcery", matCampusRenovationOracle, nil)
	passPriorityAroundTable(t, g)
	if got := lib - me.Library.Size(); got != 2 {
		t.Errorf("exiled %d cards, want 2 even with nothing to return", got)
	}
}

func TestMatB02CampusRenovationRefusesACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	bear := wthGraveyardCard(me, "Bear", wthBearType)
	if err := castCatalogSpellErr(t, g, "Campus Renovation", "Sorcery", matCampusRenovationOracle, cardTarget(bear)); err == nil {
		t.Fatal("targeted a creature card")
	}
}

// --- Harnessed Snubhorn --------------------------------------------------

func TestMatB02HarnessedSnubhornReturnsAnEnchantmentOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	horn := pushDiesCreatureForTest(g, me.ID, "Harnessed Snubhorn", matHarnessedSnubhornOracle, "Creature — Dinosaur", 2, 5)
	ench := wthGraveyardCard(me, "Old Ward", "Enchantment")
	wthGraveyardCard(me, "Dead Bear", wthBearType)
	attackWith(t, g, opp.ID, horn)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, ench) {
			pickCard(t, g, me.ID, ench)
		}
	}
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "Old Ward"); !ok {
		t.Error("the enchantment card did not return to the battlefield")
	}
	if _, ok := visBattlefieldNamed(g, me.ID, "Dead Bear"); ok {
		t.Error("a creature card came back")
	}
}

// --- Death-Rattle Oni ----------------------------------------------------

func TestMatB02DeathRattleOniDestroysOnlyDamagedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hurt := b12Creature(g, opp.ID, "Hurt Bear", wthBearType, 3, 3)
	fresh := b12Creature(g, opp.ID, "Fresh Bear", wthBearType, 3, 3)
	mine := b12Creature(g, me.ID, "My Hurt Bear", wthBearType, 3, 3)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, mine, 1)
	})
	castCatalogSpell(t, g, "Death-Rattle Oni", "Creature — Demon Spirit", matDeathRattleOniOracle, nil)
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, opp, hurt); zone != "graveyard" {
		t.Errorf("the damaged creature is in %s, want the graveyard", zone)
	}
	if zone, _ := cardWhere(g, me, mine); zone != "graveyard" {
		t.Errorf("your damaged creature is in %s, want the graveyard", zone)
	}
	if zone, _ := cardWhere(g, opp, fresh); zone != "battlefield" {
		t.Errorf("the undamaged creature is in %s, want the battlefield", zone)
	}
	if _, ok := visBattlefieldNamed(g, me.ID, "Death-Rattle Oni"); !ok {
		t.Error("the Oni destroyed itself: \"other\" creatures only")
	}
}

func TestMatB02DeathRattleOniCostsTwoLessPerCreatureThatDied(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	a := b12Creature(g, me.ID, "A", wthBearType, 1, 1)
	b := b12Creature(g, me.ID, "B", wthBearType, 1, 1)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, a, 5)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, b, 5)
	})
	g.RunStateChecksForTest()
	if n := b11CreaturesDiedThisTurn(g); n != 2 {
		t.Fatalf("%d creatures died, want 2", n)
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Death-Rattle Oni", TypeLine: "Creature — Demon Spirit",
		OracleID: matDeathRattleOniOracle, ManaCost: "{6}{B}", Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	fillPoolColored(me, "B", 1)
	fillPool(me, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("two creatures died, so {6}{B} costs {2}{B}, but the cast failed with {B}+2: %v", err)
	}
}

func TestMatB02DeathRattleOniIsFullPriceWhenNothingDied(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Death-Rattle Oni", TypeLine: "Creature — Demon Spirit",
		OracleID: matDeathRattleOniOracle, ManaCost: "{6}{B}", Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	fillPoolColored(me, "B", 1)
	fillPool(me, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast a {6}{B} creature for three mana with nothing having died")
	}
}

// --- Feast of the Victorious Dead ----------------------------------------

func TestMatB02FeastGainsLifeAndDistributesAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Feast of the Victorious Dead", "Enchantment", matFeastOracle, 0, 0)
	a := b12Creature(g, me.ID, "Bear A", wthBearType, 2, 2)
	b := b12Creature(g, me.ID, "Bear B", wthBearType, 2, 2)
	c := b12Creature(g, me.ID, "Bear C", wthBearType, 2, 2)
	doomed := b12Creature(g, g.Seats[1].ID, "Doomed", wthBearType, 1, 1)
	other := b12Creature(g, g.Seats[1].ID, "Doomed Two", wthBearType, 1, 1)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, doomed, 5)
		_ = g.DealDamageToCreatureForEffect(uuid.Nil, other, 5)
	})
	g.RunStateChecksForTest()
	life := me.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	// Two creatures died: pick two of the three, then name the first's share.
	answerChooseCards(t, g, me.ID, a, b)
	answerNumber(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if n := counterCount(g, a, "+1/+1"); n != 1 {
		t.Errorf("A has %d counters, want the 1 named", n)
	}
	if n := counterCount(g, b, "+1/+1"); n != 1 {
		t.Errorf("B has %d counters, want the 1 left over", n)
	}
	if n := counterCount(g, c, "+1/+1"); n != 0 {
		t.Errorf("C has %d counters, want none", n)
	}
}

func TestMatB02FeastOneCreatureTakesEveryCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Feast of the Victorious Dead", "Enchantment", matFeastOracle, 0, 0)
	a := b12Creature(g, me.ID, "Bear A", wthBearType, 2, 2)
	b12Creature(g, me.ID, "Bear B", wthBearType, 2, 2)
	victims := []uuid.UUID{
		b12Creature(g, g.Seats[1].ID, "V1", wthBearType, 1, 1),
		b12Creature(g, g.Seats[1].ID, "V2", wthBearType, 1, 1),
		b12Creature(g, g.Seats[1].ID, "V3", wthBearType, 1, 1),
	}
	g.WithWriteLock(func() {
		for _, v := range victims {
			_ = g.DealDamageToCreatureForEffect(uuid.Nil, v, 5)
		}
	})
	g.RunStateChecksForTest()
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, a)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, a, "+1/+1"); n != 3 {
		t.Errorf("A has %d counters, want all 3", n)
	}
}

func TestMatB02FeastDoesNothingWhenNothingDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Feast of the Victorious Dead", "Enchantment", matFeastOracle, 0, 0)
	a := b12Creature(g, me.ID, "Bear A", wthBearType, 2, 2)
	life := me.Life
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil || me.Life != life || counterCount(g, a, "+1/+1") != 0 {
		t.Error("the trigger did something with no creature having died")
	}
}

// --- Undercity Upheaval --------------------------------------------------

func TestMatB02UndercityUpheavalDistributesByCreatureCardsAndGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	a := b12Creature(g, me.ID, "Bear A", wthBearType, 2, 2)
	b := b12Creature(g, me.ID, "Bear B", wthBearType, 2, 2)
	for i := 0; i < 3; i++ {
		wthGraveyardCard(me, "Dead", wthBearType)
	}
	wthGraveyardCard(me, "Not a creature", "Sorcery")
	d1658Cast(t, g, "Undercity Upheaval", "Sorcery", matUndercityUpheavalOracle, game.CastSpellParams{
		Targets: cardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2},
	})
	passPriorityAroundTable(t, g)
	if n := counterCount(g, a, "+1/+1"); n != 1 {
		t.Errorf("A has %d counters, want 1", n)
	}
	if n := counterCount(g, b, "+1/+1"); n != 2 {
		t.Errorf("B has %d counters, want 2", n)
	}
	if !hasEffectiveKeyword(t, g, a, "vigilance") || !hasEffectiveKeyword(t, g, b, "vigilance") {
		t.Error("creatures you control did not gain vigilance")
	}
}

func TestMatB02UndercityUpheavalRefusesASplitThatIsNotX(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	a := b12Creature(g, me.ID, "Bear A", wthBearType, 2, 2)
	for i := 0; i < 3; i++ {
		wthGraveyardCard(me, "Dead", wthBearType)
	}
	err := d1658CastErr(t, g, "Undercity Upheaval", "Sorcery", matUndercityUpheavalOracle, game.CastSpellParams{
		Targets: cardRefs(a), Distribution: map[uuid.UUID]int{a: 2},
	})
	if err == nil {
		t.Fatal("divided 2 counters with 3 creature cards in the graveyard")
	}
}

func TestMatB02UndercityUpheavalWithNoCreatureCardsOnlyGrantsVigilance(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	a := b12Creature(g, me.ID, "Bear A", wthBearType, 2, 2)
	d1658Cast(t, g, "Undercity Upheaval", "Sorcery", matUndercityUpheavalOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if counterCount(g, a, "+1/+1") != 0 || !hasEffectiveKeyword(t, g, a, "vigilance") {
		t.Error("X was 0: want vigilance and no counters")
	}
}

// --- Tolarian Contempt ---------------------------------------------------

func TestMatB02TolarianContemptCountersThenTucksAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b12Creature(g, opp.ID, "Their Bear", wthBearType, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", wthBearType, 2, 2)
	castCatalogSpell(t, g, "Tolarian Contempt", "Enchantment", matTolarianContemptOracle, nil)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, theirs, "rejection"); n != 1 {
		t.Fatalf("the opposing creature has %d rejection counters, want 1", n)
	}
	if n := counterCount(g, mine, "rejection"); n != 0 {
		t.Errorf("your creature has %d rejection counters, want 0", n)
	}
	advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, mine) {
			t.Error("a creature without a rejection counter is offered")
		}
		pickCard(t, g, me.ID, theirs)
	}
	passPriorityAroundTable(t, g)
	// Its owner chooses the top or the bottom.
	pick := latestChoiceOfKindFor(g, game.PendingChoicePutInLibrary, opp.ID)
	if pick == nil {
		t.Fatalf("the owner was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolvePutInLibrary(pick.ID, opp.ID, []uuid.UUID{theirs}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if !opp.Library.Contains(theirs) {
		t.Error("the creature with a rejection counter is not in its owner's library")
	}
}

func TestMatB02TolarianContemptSkipsCreaturesWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	b12Push(g, me.ID, "Tolarian Contempt", "Enchantment", matTolarianContemptOracle, 0, 0)
	late := b12Creature(g, opp.ID, "Late Bear", wthBearType, 2, 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil && hasID(p.PickTargetCards, late) {
		t.Error("a creature that never got a rejection counter is a legal target")
	}
	if zone, _ := cardWhere(g, opp, late); zone != "battlefield" {
		t.Errorf("the creature is in %s, want the battlefield", zone)
	}
}

// --- Tranquil Frillback --------------------------------------------------

func matFrillbackEnters(t *testing.T, g *game.Game, mana int) (me, opp *game.Player) {
	t.Helper()
	advanceToMain(t, g)
	me = top100ActiveSeat(g)
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fillPoolColored(me, "G", mana)
	castCatalogSpell(t, g, "Tranquil Frillback", "Creature — Dinosaur", matTranquilFrillbackOracle, nil)
	passPriorityAroundTable(t, g)
	return me, opp
}

func TestMatB02TranquilFrillbackPayingOnceGainsFourLife(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := matFrillbackEnters(t, g, 3)
	pa := answerNumber(t, g, me.ID, 1)
	if pa.Max != 3 {
		t.Errorf("up to %d payments offered, want 3", pa.Max)
	}
	if c := answerPayUnless(t, g, me.ID, true); c.PayCost != "{G}" {
		t.Fatalf("paid %s, want {G}", c.PayCost)
	}
	life := me.Life
	c := modePickChoiceFor(g, me.ID)
	if c == nil || c.ModeMax != 1 {
		t.Fatalf("mode prompt %+v, want up to one bullet", c)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{2}); err != nil {
		t.Fatalf("gain life bullet: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestMatB02TranquilFrillbackPaidThriceRunsAllThreeBullets(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := top100ActiveSeat(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	grave := wthGraveyardCard(opp, "Their Dead", wthBearType)
	fillPoolColored(me, "G", 3)
	castCatalogSpell(t, g, "Tranquil Frillback", "Creature — Dinosaur", matTranquilFrillbackOracle, nil)
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 3)
	answerPayUnless(t, g, me.ID, true)
	c := modePickChoiceFor(g, me.ID)
	if c == nil || c.ModeMax != 3 || len(c.ModeOptionIndex) != 3 {
		t.Fatalf("mode prompt %+v, want 3 of 3", c)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0, 1, 2}); err != nil {
		t.Fatalf("all three: %v", err)
	}
	pickCard(t, g, me.ID, rock)
	answerPickTargetPlayer(t, g, opp.ID)
	life := me.Life
	passPriorityAroundTable(t, g)
	// Printed order: the artifact is destroyed first, then the graveyard
	// (the destroyed artifact included) is exiled.
	if zone, _ := cardWhere(g, opp, rock); zone == "battlefield" {
		t.Error("the artifact was not destroyed")
	}
	if zone, _ := cardWhere(g, opp, grave); zone != "exile" {
		t.Errorf("their graveyard card is in %s, want exile", zone)
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestMatB02TranquilFrillbackPayingNothingDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := matFrillbackEnters(t, g, 2)
	answerNumber(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	noModePick(t, g, me.ID, "nothing was paid")
	if n := me.ManaPool.Total(); n != 2 {
		t.Errorf("pool %d, want 2 untouched", n)
	}
}

// --- Nahiri's Resolve ----------------------------------------------------

func TestMatB02NahirisResolveGivesPlusOneHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Nahiri's Resolve", "Enchantment", matNahirisResolveOracle, 0, 0)
	mine := b12Creature(g, me.ID, "My Bear", wthBearType, 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Their Bear", wthBearType, 2, 2)
	if p, _ := ptOf(t, g, mine); p != 3 || !hasEffectiveKeyword(t, g, mine, "haste") {
		t.Errorf("your creature: power %d haste %v, want 3 and haste", p, hasEffectiveKeyword(t, g, mine, "haste"))
	}
	if p, _ := ptOf(t, g, theirs); p != 2 || hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Errorf("their creature should be untouched, power %d", p)
	}
}

func TestMatB02NahirisResolveExilesYourPickAndReturnsThemNextUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Nahiri's Resolve", "Enchantment", matNahirisResolveOracle, 0, 0)
	a := b12Creature(g, me.ID, "Bear A", wthBearType, 2, 2)
	b := b12Creature(g, me.ID, "Bear B", wthBearType, 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, a, rock)
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, me, a); zone != "exile" {
		t.Errorf("Bear A is in %s, want exile", zone)
	}
	if zone, _ := cardWhere(g, me, rock); zone != "exile" {
		t.Errorf("the artifact is in %s, want exile", zone)
	}
	if zone, _ := cardWhere(g, me, b); zone != "battlefield" {
		t.Errorf("Bear B is in %s, want the battlefield: it was not chosen", zone)
	}
	// Other seats' upkeeps do not return them; mine does.
	for seat := 1; seat < len(g.Seats); seat++ {
		advanceToMainOf(t, g, seat)
		if zone, _ := cardWhere(g, me, a); zone != "exile" {
			t.Fatalf("Bear A came back during seat %d's turn (%s)", seat, zone)
		}
	}
	advanceToMainOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if _, ok := visBattlefieldNamed(g, me.ID, "Bear A"); !ok {
		t.Error("Bear A did not return on my next upkeep")
	}
	if _, ok := visBattlefieldNamed(g, me.ID, "Rock"); !ok {
		t.Error("the artifact did not return on my next upkeep")
	}
}

func TestMatB02NahirisResolveMayExileNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Nahiri's Resolve", "Enchantment", matNahirisResolveOracle, 0, 0)
	a := b12Creature(g, me.ID, "Bear A", wthBearType, 2, 2)
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if zone, _ := cardWhere(g, me, a); zone != "battlefield" {
		t.Errorf("Bear A is in %s, want the battlefield", zone)
	}
}
