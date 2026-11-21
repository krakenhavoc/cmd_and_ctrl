package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	taintedAdversaryOracle  = "8bed92cc-b3f8-4201-bcbc-44535ff1bd96"
	primalAdversaryOracle   = "087d6330-c9de-42b7-a804-1202bc6aa682"
	intrepidAdversaryOracle = "b72fde5e-862a-45e8-9880-f2c16975a9ac"
	spectralAdversaryOracle = "0de4171e-5908-408e-9852-25f148d0fb47"
	gisaOracle              = "f1354b1b-896c-4492-b0c0-3dd07c6d3917"
	consumingBlobOracle     = "818292cc-6261-47e4-b573-d92f03da6ded"
	moonveilRegentOracle    = "61c2a2cb-6424-4a20-81fb-f8e15a8f95ce"
	slogurkOracle           = "cee836f1-87a4-43db-a6ea-55575a4ce781"
	sigardasSplendorOracle  = "71a3d0ce-052b-41eb-a70f-0c716149a55f"
)

func b15Counters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[kind]
}

func TestTaintedAdversaryPaysTwiceForTwoCountersAndFourDecayedZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 2)
	effectstest.FillPool(me, 8)
	adv := effectstest.CastHoldCreature(t, g, "Tainted Adversary", "Creature — Zombie", taintedAdversaryOracle, 2, 3)
	pa := effectstest.AnswerNumber(t, g, me.ID, 2)
	if pa.Min != 0 {
		t.Fatalf("X prompt = %+v, want 0 up", pa)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b15Counters(t, g, adv, game.CounterPlusOne); n != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", n)
	}
	zs := effectstest.TokensNamed(g, "Zombie")
	if len(zs) != 4 {
		t.Fatalf("%d Zombies, want twice 2 = 4", len(zs))
	}
	for _, z := range zs {
		if !effectstest.HasEffectiveKeyword(t, g, z.InstanceID, game.KeywordDecayed) || z.Controller != me.ID {
			t.Errorf("Zombie %v is not a decayed Zombie of mine", z.InstanceID)
		}
	}
}

func TestTaintedAdversaryPayingNothingMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 2)
	effectstest.FillPool(me, 6)
	adv := effectstest.CastHoldCreature(t, g, "Tainted Adversary", "Creature — Zombie", taintedAdversaryOracle, 2, 3)
	effectstest.AnswerNumber(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b15Counters(t, g, adv, game.CounterPlusOne); n != 0 {
		t.Errorf("counters = %d, want none", n)
	}
	if zs := effectstest.TokensNamed(g, "Zombie"); len(zs) != 0 {
		t.Errorf("%d Zombies, want none", len(zs))
	}
}

func TestPrimalAdversaryTurnsUpToThatManyLandsIntoHastyWolvesForGood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	forest := effectstest.BfCardWithOracle(g, me, "Forest", "Basic Land — Forest", "")
	other := effectstest.BfCardWithOracle(g, me, "Mountain", "Basic Land — Mountain", "")
	theirs := effectstest.BfCardWithOracle(g, opp, "Island", "Basic Land — Island", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 2)
	effectstest.FillPool(me, 4)
	adv := effectstest.CastHoldCreature(t, g, "Primal Adversary", "Creature — Wolf", primalAdversaryOracle, 4, 3)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	ch := effectstest.PickTargetPrompt(t, g)
	if len(ch.PickTargetCards) != 2 || ch.PickTargetMax != 2 || ch.PickTargetMin != 0 {
		t.Fatalf("up to two lands I control: %v (%d..%d)", ch.PickTargetCards, ch.PickTargetMin, ch.PickTargetMax)
	}
	for _, id := range ch.PickTargetCards {
		if id == theirs {
			t.Fatal("an opponent's land is offered")
		}
	}
	effectstest.AnswerPickTarget(t, g, forest)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b15Counters(t, g, adv, game.CounterPlusOne); n != 2 {
		t.Errorf("counters = %d, want 2", n)
	}
	p, tough := effectstest.EffectivePTOf(t, g, forest)
	if p != 3 || tough != 3 {
		t.Errorf("the land is %d/%d, want 3/3", p, tough)
	}
	c, _ := effectstest.BattlefieldCard(g, forest)
	eff := c.Effective()
	if !slices.Contains(eff.Types, "Creature") || !slices.Contains(eff.Subtypes, "Wolf") ||
		!slices.Contains(eff.Types, "Land") || !slices.Contains(eff.Subtypes, "Forest") {
		t.Errorf("types %v %v: want a Wolf creature that is still a Forest land", eff.Types, eff.Subtypes)
	}
	if !effectstest.HasEffectiveKeyword(t, g, forest, "haste") {
		t.Error("the Wolf land has no haste")
	}
	if c2, _ := effectstest.BattlefieldCard(g, other); slices.Contains(c2.Effective().Types, "Creature") {
		t.Error("an unpicked land became a creature")
	}
	// No duration: it is still a Wolf after the turn passes.
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, forest); p != 3 {
		t.Errorf("the land stopped being a 3/3 after the turn ended (%d)", p)
	}
}

func TestIntrepidAdversaryValorCountersPumpEveryCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	mine := effectstest.BfCardWithOracle(g, me, "Bear", "Creature — Bear", "")
	theirs := effectstest.BfCardWithOracle(g, opp, "Bear", "Creature — Bear", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "W", 2)
	effectstest.FillPool(me, 4)
	adv := effectstest.CastHoldCreature(t, g, "Intrepid Adversary", "Creature — Human Scout", intrepidAdversaryOracle, 3, 1)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b15Counters(t, g, adv, "valor"); n != 2 {
		t.Fatalf("valor counters = %d, want 2", n)
	}
	if p := effectstest.EffectivePower(t, g, mine); p != 3 {
		t.Errorf("my 1/1 has power %d, want 3", p)
	}
	if p := effectstest.EffectivePower(t, g, adv); p != 5 {
		t.Errorf("the Adversary has power %d, want 3 plus 2", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("an opponent's creature has power %d, want 1", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, adv, "lifelink") {
		t.Error("no lifelink")
	}
}

func TestSpectralAdversaryPhasesOutOtherPermanentsButNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.BfCardWithOracle(g, opp, "Bear", "Creature — Bear", "")
	land := effectstest.BfCardWithOracle(g, opp, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "U", 2)
	effectstest.FillPool(me, 4)
	adv := effectstest.CastHoldCreature(t, g, "Spectral Adversary", "Creature — Spirit", spectralAdversaryOracle, 2, 1)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, id := range ch.PickTargetCards {
		if id == adv {
			t.Fatal("the Adversary itself is offered: 'other'")
		}
		if id == land {
			t.Fatal("a land is offered: artifacts, creatures and enchantments only")
		}
	}
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := b15Counters(t, g, adv, game.CounterPlusOne); n != 2 {
		t.Errorf("counters = %d, want 2", n)
	}
	if !effectstest.PhasedOutInCatalogGame(g, bear) {
		t.Error("the targeted creature did not phase out")
	}
	if effectstest.PhasedOutInCatalogGame(g, land) || effectstest.PhasedOutInCatalogGame(g, adv) {
		t.Error("the land or the Adversary phased out")
	}
}

func TestGisaExilesAnOpponentsDyingCreatureAndReturnsItDecayedUnderYourControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	gisa := effectstest.B12Push(g, me.ID, "Gisa, Glorious Resurrector", "Legendary Creature — Human Wizard", gisaOracle, 4, 4)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	mine := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(theirs)
		_ = g.DestroyPermanentForEffect(mine)
	})
	if opp.Graveyard.Contains(theirs) || !g.Exile.Contains(theirs) {
		t.Fatal("the opponent's creature did not go to exile instead of dying")
	}
	if !me.Graveyard.Contains(mine) {
		t.Fatal("my own creature should die as usual")
	}
	var link game.PermissionCardRef
	for _, c := range g.Exile.Cards {
		if c.InstanceID == theirs {
			link = c.ExiledWith
		}
	}
	if link.ID != gisa {
		t.Fatalf("exiled with %v, want Gisa", link.ID)
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	bears := effectstest.BattlefieldIDsNamed(g, "Their Bear")
	if len(bears) != 1 || effectstest.ControllerOf(t, g, bears[0]) != me.ID {
		t.Fatalf("the exiled creature is not on my battlefield: %v", bears)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bears[0], game.KeywordDecayed) {
		t.Error("the returned creature has no decayed")
	}
	if g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Error("still in exile or the graveyard")
	}
}

func TestGisaDoesNotReturnCreaturesExiledByAnotherGisa(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%4]
	gisa := effectstest.B12Push(g, me.ID, "Gisa, Glorious Resurrector", "Legendary Creature — Human Wizard", gisaOracle, 4, 4)
	foreign := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: foreign, Name: "Elsewhere", TypeLine: "Creature — Bear", Power: 1, Toughness: 1,
		Owner: opp.ID, Controller: opp.ID, ExiledWith: game.PermissionCardRef{ID: uuid.New(), Epoch: 1}})
	_ = gisa
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.BattlefieldIDsNamed(g, "Elsewhere")) != 0 {
		t.Error("a creature not exiled with Gisa came back")
	}
}

func TestConsumingBlobCountsCardTypesInYourGraveyardAndMakesASelfSizingOoze(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	blob := effectstest.B12Push(g, me.ID, "Consuming Blob", "Creature — Ooze", consumingBlobOracle, 0, 0)
	if p, tough := effectstest.EffectivePTOf(t, g, blob); p != 0 || tough != 1 {
		t.Fatalf("with an empty graveyard the Blob is %d/%d, want 0/1", p, tough)
	}
	effectstest.GyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.GyCard(me, "Forest", "Basic Land — Forest", "", "")
	effectstest.GyCard(me, "Other Bear", "Creature — Bear", "{1}{G}", "")
	g.BumpLayerVersionForTest()
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.EffectivePTOf(t, g, blob); p != 2 || tough != 3 {
		t.Fatalf("creature and land types: the Blob is %d/%d, want 2/3", p, tough)
	}
	effectstest.GyCard(me, "Artifact Creature", "Artifact Creature — Golem", "{3}", "")
	g.BumpLayerVersionForTest()
	if p, tough := effectstest.EffectivePTOf(t, g, blob); p != 3 || tough != 4 {
		t.Fatalf("creature, land and artifact: the Blob is %d/%d, want 3/4", p, tough)
	}
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	oozes := effectstest.TokensNamed(g, "Ooze")
	if len(oozes) != 1 || oozes[0].Controller != me.ID {
		t.Fatalf("%d Oozes, want one of mine", len(oozes))
	}
	ooze := oozes[0].InstanceID
	if p, tough := effectstest.EffectivePTOf(t, g, ooze); p != 3 || tough != 4 {
		t.Errorf("the Ooze is %d/%d, want 3/4", p, tough)
	}
	effectstest.GyCard(me, "Sorcery", "Sorcery", "{1}", "")
	g.BumpLayerVersionForTest()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(blob) })
	if p, tough := effectstest.EffectivePTOf(t, g, ooze); p != 4 || tough != 5 {
		t.Errorf("after a sorcery and the Blob dying the Ooze is %d/%d, want 4/5", p, tough)
	}
}

func TestMoonveilRegentDiscardsYourHandToDrawPerColorOfTheSpellYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Moonveil Regent", "Creature — Dragon", moonveilRegentOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.HandCardFull(me, "Filler", "Sorcery", "{9}", "", nil)
	spell := effectstest.HandCardFull(me, "Azorius Charm", "Instant", "", "", []string{"W", "U"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if c == nil {
		t.Fatal("no may-discard prompt")
	}
	handBefore := me.Hand.Size()
	if handBefore < 2 {
		t.Fatalf("hand = %d, want a hand to discard", handBefore)
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if got := me.Hand.Size(); got != 2 {
		t.Errorf("hand after discarding it and drawing one per color = %d, want 2", got)
	}
}

func TestMoonveilRegentDecliningKeepsYourHandAndColorlessSpellsDrawNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Moonveil Regent", "Creature — Dragon", moonveilRegentOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.HandCardFull(me, "Colorless Thing", "Instant", "", "", nil)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if c == nil {
		t.Fatal("no may-discard prompt")
	}
	before := me.Hand.Size()
	if err := g.ResolveConfirm(c.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != before {
		t.Errorf("declining changed the hand: %d -> %d", before, me.Hand.Size())
	}
}

func TestMoonveilRegentDiesForXDamageWhereXIsTheColorsAmongYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	regent := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Moonveil Regent",
		TypeLine: "Creature — Dragon", OracleID: moonveilRegentOracle, Power: 4, Toughness: 4, Colors: []string{"R"},
		Owner: me.ID, Controller: me.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Thing", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Colors: []string{"W"}, Owner: me.ID, Controller: me.ID})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Blue Thing", TypeLine: "Creature — Test",
		Power: 1, Toughness: 1, Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(regent) })
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	before := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d -> %d, want 2 damage (white and blue; the Regent is gone)", before, opp.Life)
	}
}

func TestSlogurkGrowsForEachLandCardPutIntoYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	slogurk := effectstest.B12Push(g, me.ID, "Slogurk, the Overslime", "Legendary Creature — Ooze", slogurkOracle, 3, 3)
	land := effectstest.BfCardWithOracle(g, me, "Forest", "Basic Land — Forest", "")
	theirLand := effectstest.BfCardWithOracle(g, opp, "Island", "Basic Land — Island", "")
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(land)
		_ = g.DestroyPermanentForEffect(theirLand)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := b15Counters(t, g, slogurk, game.CounterPlusOne); n != 1 {
		t.Errorf("counters = %d, want 1 (only my land card counts)", n)
	}
}

func TestSlogurkRemovesThreeCountersToReturnItselfThenFetchesUpToThreeLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	slogurk := effectstest.B12Push(g, me.ID, "Slogurk, the Overslime", "Legendary Creature — Ooze", slogurkOracle, 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == slogurk {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2}
			}
		}
	})
	l1 := effectstest.GyCard(me, "Forest", "Basic Land — Forest", "", "")
	l2 := effectstest.GyCard(me, "Island", "Basic Land — Island", "", "")
	bear := effectstest.GyCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	if err := g.ActivateCatalogAbility(me.ID, slogurk, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only two counters to remove")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == slogurk {
				g.Battlefield.Cards[i].Counters[game.CounterPlusOne] = 3
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, slogurk, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(slogurk) {
		t.Fatal("Slogurk was not returned to its owner's hand")
	}
	ch := effectstest.PickTargetPrompt(t, g)
	for _, id := range ch.PickTargetCards {
		if id == bear {
			t.Fatal("a non-land card is offered")
		}
	}
	if ch.PickTargetMax != 3 || ch.PickTargetMin != 0 {
		t.Fatalf("up to three: %d..%d", ch.PickTargetMin, ch.PickTargetMax)
	}
	effectstest.AnswerPickTarget(t, g, l1)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(l1) {
		t.Error("the picked land did not return to hand")
	}
	if !me.Graveyard.Contains(l2) || !me.Graveyard.Contains(bear) {
		t.Error("an unpicked card left the graveyard")
	}
}

func TestSigardasSplendorNotesYourLifeAndDrawsOnlyWhenItHasNotDropped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	me.Life = 31
	effectstest.AdvanceToMain(t, g)
	splendor := effectstest.CastCatalogSpell(t, g, "Sigarda's Splendor", "Enchantment", sigardasSplendorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := g.ChosenNumberOf(splendor); n != 31 {
		t.Fatalf("noted life total = %d, want 31", n)
	}
	// Life unchanged: the upkeep draws, then notes again.
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	before := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Fatalf("hand %d -> %d, want one drawn at an unchanged life total", before, me.Hand.Size())
	}
	// Life dropped: no draw, and the lower total is now the note.
	me.Life = 20
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	before = me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Fatalf("hand %d -> %d, want no draw after losing life", before, me.Hand.Size())
	}
	if n := g.ChosenNumberOf(splendor); n != 20 {
		t.Errorf("noted life total = %d, want 20", n)
	}
	// Life gained: draws again.
	me.Life = 25
	effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
	before = me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d -> %d, want a draw after gaining life", before, me.Hand.Size())
	}
}

func TestNotedNumberKeepsANegativeLifeTotalThatTheChosenNumberWouldClamp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	splendor := effectstest.PushPermanentForTest(g, me.ID, "Sigarda's Splendor", sigardasSplendorOracle, "Enchantment")
	g.WithWriteLock(func() { g.SetNotedNumberForEffect(splendor, -5) })
	if n := g.ChosenNumberOf(splendor); n != -5 {
		t.Fatalf("noted number = %d, want -5", n)
	}
	g.WithWriteLock(func() { g.SetChosenNumberForEffect(splendor, -5) })
	if n := g.ChosenNumberOf(splendor); n != 0 {
		t.Fatalf("the chosen-number setter's clamp is gone: %d", n)
	}
}

func TestSigardasSplendorGainsALifeForEachWhiteSpellYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushPermanentForTest(g, me.ID, "Sigarda's Splendor", sigardasSplendorOracle, "Enchantment")
	white := effectstest.HandCardFull(me, "White Instant", "Instant", "", "", []string{"W"})
	green := effectstest.HandCardFull(me, "Green Instant", "Instant", "", "", []string{"G"})
	life := me.Life
	if err := g.CastSpell(me.ID, green, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast green: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Fatalf("a green spell gained life: %d -> %d", life, me.Life)
	}
	if err := g.CastSpell(me.ID, white, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast white: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}
