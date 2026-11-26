package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b301_cards_test.go — Oath of the Gatewatch, batch 301 (S82 older
// sets, #5545). Every test plays the real catalog entry through the engine.

const (
	oldB301CrumblingVestige = "3c1fe86a-3696-4634-bc26-cd45846529ec"
	oldB301Remorseless      = "42e07167-b490-46ad-9b5c-b483a4bb64e3"
	oldB301BalothNull       = "55a0a40c-cb40-469d-9bf3-9c33b378f4f7"
	oldB301SeedGuardian     = "728045bf-0959-44ff-8de9-16ec35fd248f"
	oldB301Nissa            = "b0f36bf7-bdde-4576-9eca-ec02d41326ab"
	oldB301GiftOfTusks      = "c0307f8d-b39a-4180-8a45-ddc019ee39f6"
)

func TestOldB301EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{oldB301CrumblingVestige, oldB301Remorseless, oldB301BalothNull, oldB301SeedGuardian, oldB301Nissa, oldB301GiftOfTusks} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestOldB301CrumblingVestigeEntersTappedAndAddsOneManaOfAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PlayLandFromHand(t, g, "Crumbling Vestige", oldB301CrumblingVestige)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.RiderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("the enters trigger offers any colour: %+v", pick)
	}
	effectstest.RiderAnswerManaPicks(t, g, me.ID, "U")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want {U}", got)
	}
	if !effectstest.B16Tapped(t, g, id) {
		t.Error("the land entered untapped")
	}
}

func TestOldB301CrumblingVestigeTapsForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Crumbling Vestige", "Land", oldB301CrumblingVestige, false)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want {C}", got)
	}
}

func oldB301Creature(p *game.Player, name string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Creature — Bear"
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

func TestOldB301BalothNullReturnsUpToTwoCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := oldB301Creature(me, "Dead Bear A")
	b := oldB301Creature(me, "Dead Bear B")
	c := oldB301Creature(me, "Dead Bear C")
	land := game.NewCard("Dead Land", me.ID)
	land.TypeLine = "Land"
	me.Graveyard.PushTop(land)
	effectstest.CastHoldCreature(t, g, "Baloth Null", "Creature — Zombie Beast", oldB301BalothNull, 4, 5)
	pc := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(pc.ID, pc.Chooser, effectstest.CardRefs(a, b)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) {
		t.Errorf("the two chosen creature cards did not return to hand")
	}
	if me.Hand.Contains(c) || !me.Graveyard.Contains(c) {
		t.Error("an unchosen creature card left the graveyard")
	}
	if me.Hand.Contains(land.InstanceID) {
		t.Error("a land card returned")
	}
}

func TestOldB301SeedGuardianLeavesAnElementalAsBigAsTheCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB301Creature(me, "Dead Bear A")
	oldB301Creature(me, "Dead Bear B")
	land := game.NewCard("Dead Land", me.ID)
	land.TypeLine = "Land"
	me.Graveyard.PushTop(land)
	guard := effectstest.B12Push(g, me.ID, "Seed Guardian", "Creature — Elemental", oldB301SeedGuardian, 3, 4)
	effectstest.KillCreature(t, g, me.ID, guard)
	effectstest.PassPriorityAroundTable(t, g)
	found := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Elemental" && c.Controller == me.ID {
			found++
			if c.Power != 3 || c.Toughness != 3 || !c.HasColor("G") {
				t.Errorf("Elemental is %d/%d, want 3/3 green (two Bears and the Guardian itself)", c.Power, c.Toughness)
			}
		}
	}
	if found != 1 {
		t.Errorf("%d Elemental tokens, want 1", found)
	}
}

func oldB301Walker(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	effectstest.ToMain(t, g)
	return effectstest.B12Push(g, me.ID, "Nissa, Voice of Zendikar", "Legendary Planeswalker — Nissa", oldB301Nissa, 0, 0)
}

func TestOldB301NissaPlusOneMakesAPlant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	walker := oldB301Walker(t, g, me)
	effectstest.B16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	found := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsToken() && c.Name == "Plant" && c.Controller == me.ID {
			found++
			if c.Power != 0 || c.Toughness != 1 || !c.HasColor("G") {
				t.Errorf("Plant is %d/%d", c.Power, c.Toughness)
			}
		}
	}
	if found != 1 {
		t.Errorf("%d Plant tokens, want 1", found)
	}
}

func TestOldB301NissaMinusTwoPutsACounterOnEachCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := oldB301Walker(t, g, me)
	// Loyalty: the pushed walker has none; the ability needs 2.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(walker, game.CounterLoyalty, 3) })
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	mine2 := effectstest.B12Creature(g, me.ID, "Mine Too", "Creature — Bear", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOneCounters(g, mine) != 1 || effectstest.PlusOneCounters(g, mine2) != 1 {
		t.Error("each creature you control should get a +1/+1 counter")
	}
	if effectstest.PlusOneCounters(g, theirs) != 0 {
		t.Error("an opponent's creature got a counter")
	}
}

func TestOldB301NissaMinusSevenGainsAndDrawsForEachLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	walker := oldB301Walker(t, g, me)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(walker, game.CounterLoyalty, 7) })
	for i := 0; i < 4; i++ {
		effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	}
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	life, hand := effectstest.LifeOf(g, me.ID), me.Hand.Size()
	effectstest.B16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID) - life; got != 4 {
		t.Errorf("gained %d life, want 4", got)
	}
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d cards, want 4", got)
	}
}

func TestOldB301GiftOfTusksMakesAThreeThreeGreenElephantWithNoAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Push(g, opp.ID, "Their Dragon", "Creature — Dragon", "", 5, 5)
	g.WithWriteLock(func() {
		if c, ok := g.LookupCardForEffect(bear); ok {
			c.Keywords = append(c.Keywords, "flying")
		}
	})
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, "Gift of Tusks", "Instant", "", oldB301GiftOfTusks)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.JmpB4Live(t, g, bear)
	if live.CurrentPower() != 3 || live.CurrentToughness() != 3 || !live.HasColor("G") || len(live.EffectiveColors()) != 1 {
		t.Errorf("target is %d/%d colours %v, want a green 3/3", live.CurrentPower(), live.CurrentToughness(), live.EffectiveColors())
	}
	if !live.HasSubtype("Elephant") || live.HasSubtype("Dragon") {
		t.Errorf("target type line %q, want an Elephant", live.TypeLine)
	}
	if game.HasKeyword(live, "flying") {
		t.Error("the target kept flying")
	}
}

func TestOldB301GiftOfTusksWearsOffAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, "Gift of Tusks", "Instant", "", oldB301GiftOfTusks)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if live := effectstest.JmpB4Live(t, g, bear); live.CurrentPower() != 3 {
		t.Fatalf("not 3/3 this turn: %d", live.CurrentPower())
	}
	effectstest.EndTurn(t, g)
	if live := effectstest.JmpB4Live(t, g, bear); live.CurrentPower() != 2 || live.CurrentToughness() != 2 {
		t.Errorf("after the turn the bear is %d/%d, want 2/2", live.CurrentPower(), live.CurrentToughness())
	}
}

// oldB301Punish casts Remorseless Punishment from the active seat at opp.
func oldB301Punish(t *testing.T, g *game.Game, me, opp *game.Player) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	effectstest.D5681Spell(g, me, id, "Remorseless Punishment", "Sorcery", "", oldB301Remorseless)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB301RemorselessPunishmentOffersThreeWaysAndRepeatsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	walker := effectstest.PushWalkerForCounterCost(g, opp.ID, 3)
	life := opp.Life
	oldB301Punish(t, g, me, opp)
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil || len(pick.PickOptions) != 3 {
		t.Fatalf("lose life / sacrifice / discard expected: %+v", pick)
	}
	if pick.PickOptions[0].LifeCost != 5 {
		t.Errorf("the life branch declares its price: %+v", pick.PickOptions[0])
	}
	// First ask: sacrifice a creature or planeswalker.
	effectstest.AnswerOptionPick(t, g, opp.ID, 1)
	sac := effectstest.LatestChooseCardsFor(g, opp.ID)
	if sac == nil {
		t.Fatalf("the sacrifice branch asks which permanent: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	if opp.Life != life {
		t.Error("sacrificing costs no life")
	}
	// Second ask: the process repeats once. Now discard two.
	pick = effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the process did not repeat: %+v", g.PendingChoices)
	}
	hand := opp.Hand.Size()
	effectstest.AnswerOptionPick(t, g, opp.ID, 2)
	if got := effectstest.DiscardOwed(g, opp.ID); got != 2 {
		t.Fatalf("owes %d discards, want 2", got)
	}
	cards := opp.Hand.Cards
	effectstest.AnswerDiscard(t, g, opp.ID, cards[0].InstanceID, cards[1].InstanceID)
	if opp.Hand.Size() != hand-2 {
		t.Errorf("hand %d → %d, want -2", hand, opp.Hand.Size())
	}
	if !g.Battlefield.Contains(walker) {
		t.Error("the planeswalker should be untouched")
	}
	if effectstest.LatestOptionPickFor(g, opp.ID) != nil {
		t.Error("asked a third time")
	}
	if opp.Life != life {
		t.Errorf("life %d → %d, want no change", life, opp.Life)
	}
}

func TestOldB301RemorselessPunishmentTwoRefusalsCostTenLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	oldB301Punish(t, g, me, opp)
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	if opp.Life != life-5 {
		t.Fatalf("life %d → %d after one refusal, want -5", life, opp.Life)
	}
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	if opp.Life != life-10 {
		t.Errorf("life %d → %d after both, want -10", life, opp.Life)
	}
}

func TestOldB301RemorselessPunishmentDropsWhatThePlayerCannotDo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// One card in hand cannot be "two cards", and no creature or planeswalker
	// means nothing to sacrifice (a land and an artifact don't count).
	opp.Hand.Cards = opp.Hand.Cards[:1]
	effectstest.B12Permanent(g, opp.ID, "Their Land", "Basic Land — Forest")
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	oldB301Punish(t, g, me, opp)
	pick := effectstest.LatestOptionPickFor(g, opp.ID)
	if pick == nil || len(pick.PickOptions) != 1 || pick.PickOptions[0].Label != "Lose 5 life" {
		t.Fatalf("only the life branch is possible: %+v", pick)
	}
}
