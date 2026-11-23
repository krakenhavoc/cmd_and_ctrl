package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestE2CommanderBatchRegisters(t *testing.T) {
	want := map[string]string{
		effectstest.E2BetorOracle:            "Betor, Ancestor's Voice",
		effectstest.E2RodolfOracle:           "Rodolf Duskbringer",
		effectstest.E2EnduringTenacityOracle: "Enduring Tenacity",
		effectstest.E2KrrikOracle:            "K'rrik, Son of Yawgmoth",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
	}
}

// e2CountersOn reads a permanent's +1/+1 counters.
func e2CountersOn(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterPlusOne]
			}
		}
	})
	return n
}

// --- Betor, Ancestor's Voice --------------------------------------

// The commander's end step: counters equal to the life gained, and a
// reanimation bounded by the life lost. Both clauses target "up to
// one" and belong to ONE trigger, so the pick asks the battlefield
// clause first and the graveyard clause second, and there is nothing
// to order.
func TestE2BetorCountersAndReanimatesOffTheTurnsLifeSwing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	betor := effectstest.B12Push(g, me.ID, "Betor, Ancestor's Voice", "Legendary Creature — Spirit Dragon", effectstest.E2BetorOracle, 3, 5)
	abilities := effectstest.EffectiveAbilities(t, g, betor)
	if !effectstest.HasAbility(abilities, "flying") || !effectstest.HasAbility(abilities, "lifelink") {
		t.Errorf("printed flying and lifelink: %v", abilities)
	}
	ally := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	cheap := effectstest.PushGraveyardPermanent(me, "Dead Scout", "Creature — Human Scout", "{B}")
	expensive := effectstest.PushGraveyardPermanent(me, "Dead Fatty", "Creature — Giant", "{5}{B}")
	notMine := effectstest.PushGraveyardPermanent(opp, "Their Dead Bear", "Creature — Bear", "{B}")

	// Three gained, two lost: three counters, and a reanimation
	// budget of two mana value.
	effectstest.E2ChangeLife(g, me.ID, 3)
	effectstest.E2ChangeLife(g, me.ID, -2)
	effectstest.AdvanceToEndStepOf(t, g, 0)

	// Clause 0: up to one other creature you control.
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, ally) {
		t.Fatalf("the first clause offers your other creature: %+v", p)
	}
	if p.PickTargetMin != 0 {
		t.Errorf("up to ONE other target creature you control: min %d", p.PickTargetMin)
	}
	if effectstest.HasID(p.PickTargetCards, betor) {
		t.Error("OTHER target creature — Betor may not choose itself")
	}
	if effectstest.HasID(p.PickTargetCards, theirs) {
		t.Error("a creature you control — not an opponent's")
	}
	if effectstest.HasID(p.PickTargetCards, cheap) {
		t.Error("the battlefield clause does not offer a graveyard card")
	}
	effectstest.PickCard(t, g, me.ID, ally)

	// Clause 1, same trigger: up to one creature card in your
	// graveyard within the life lost.
	p = effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, cheap) {
		t.Fatalf("the second clause offers the affordable creature card: %+v", p)
	}
	if effectstest.HasID(p.PickTargetCards, expensive) {
		t.Error("mana value 6 is above the two life lost this turn")
	}
	if effectstest.HasID(p.PickTargetCards, notMine) {
		t.Error("from YOUR graveyard only")
	}
	if effectstest.HasID(p.PickTargetCards, ally) {
		t.Error("the graveyard clause does not offer a creature on the battlefield")
	}
	effectstest.PickCard(t, g, me.ID, cheap)

	if n := effectstest.TriggersOnStackFrom(g, betor); n != 1 {
		t.Errorf("Betor's end step is ONE trigger on the stack, got %d", n)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if got := e2CountersOn(g, ally); got != 3 {
		t.Errorf("+1/+1 counters equal to the 3 life gained: %d", got)
	}
	if !g.Battlefield.Contains(cheap) {
		t.Error("the affordable creature card returns to the battlefield")
	}
	if !me.Graveyard.Contains(expensive) {
		t.Error("the over-budget creature card stays in the graveyard")
	}
}

// No life gained and none lost: both triggers still happen and both
// do nothing, without erroring and without leaving a prompt open.
func TestE2BetorDoesNothingWithNoLifeSwing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Betor, Ancestor's Voice", "Legendary Creature — Spirit Dragon", effectstest.E2BetorOracle, 3, 5)
	ally := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	dead := effectstest.PushGraveyardPermanent(me, "Dead Scout", "Creature — Human Scout", "{B}")

	effectstest.AdvanceToEndStepOf(t, g, 0)
	// The counters clause still targets — it is not conditional — so
	// answer it; the graveyard clause has no legal target at a budget
	// of zero and never asks.
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetCards, dead) {
		t.Fatal("nothing was lost this turn: no creature card is within the budget")
	}
	effectstest.PickCard(t, g, me.ID, ally)
	effectstest.PassPriorityAroundTable(t, g)

	if got := e2CountersOn(g, ally); got != 0 {
		t.Errorf("no life gained, no counters: %d", got)
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("no life lost, nothing reanimated")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("no prompt is left waiting")
	}
}

// --- Rodolf Duskbringer -------------------------------------------

func TestE2RodolfGainsIndestructibleThenPaysToReanimate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rodolf := effectstest.B12Push(g, me.ID, "Rodolf Duskbringer", "Legendary Creature — Vampire Angel", effectstest.E2RodolfOracle, 4, 4)
	abilities := effectstest.EffectiveAbilities(t, g, rodolf)
	for _, kw := range []string{"flying", "deathtouch", "lifelink"} {
		if !effectstest.HasAbility(abilities, kw) {
			t.Errorf("printed %s: %v", kw, abilities)
		}
	}
	if effectstest.HasAbility(abilities, "indestructible") {
		t.Error("indestructible is granted by the lifegain trigger, not printed")
	}
	cheap := effectstest.PushGraveyardPermanent(me, "Dead Scout", "Creature — Human Scout", "{1}{B}")
	expensive := effectstest.PushGraveyardPermanent(me, "Dead Fatty", "Creature — Giant", "{5}{B}")

	effectstest.E2ChangeLife(g, me.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, rodolf), "indestructible") {
		t.Fatal("gaining life grants indestructible until end of turn")
	}

	effectstest.AdvanceToEndStepOf(t, g, 0)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the end step offers the optional {1}{W/B}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)

	// "When you do" is a reflexive trigger with its own target, asked
	// after the payment.
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if !effectstest.HasID(p.PickTargetCards, cheap) {
		t.Fatalf("mana value 2 is within the 5 life gained: %+v", p)
	}
	if effectstest.HasID(p.PickTargetCards, expensive) {
		t.Error("mana value 6 is above the 5 life gained")
	}
	if p.PickTargetMin != 1 {
		t.Errorf("a required target, not an up-to-one: min %d", p.PickTargetMin)
	}
	effectstest.PickCard(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(cheap) {
		t.Error("paying returns the chosen creature card to the battlefield")
	}
	if !me.Graveyard.Contains(expensive) {
		t.Error("only the chosen card comes back")
	}
}

// Declining the optional cost raises no reflexive trigger at all.
func TestE2RodolfDecliningThePaymentReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Rodolf Duskbringer", "Legendary Creature — Vampire Angel", effectstest.E2RodolfOracle, 4, 4)
	dead := effectstest.PushGraveyardPermanent(me, "Dead Scout", "Creature — Human Scout", "{1}{B}")

	effectstest.E2ChangeLife(g, me.ID, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("declining asks for no target")
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("declining returns nothing")
	}
}

// --- Enduring Tenacity --------------------------------------------

func TestE2EnduringTenacityDrainsAndReturnsAsAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tenacity := effectstest.B12Push(g, me.ID, "Enduring Tenacity", "Enchantment Creature — Snake Glimmer", effectstest.E2EnduringTenacityOracle, 4, 3)
	before := opp.Life

	effectstest.E2ChangeLife(g, me.ID, 4)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(p.PickTargetPlayers, me.ID) {
		t.Error("target OPPONENT — not yourself")
	}
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the opponent loses the 4 life gained: %d", before-opp.Life)
	}

	// The Glimmer half: it dies as a creature and comes back as an
	// enchantment that is not a creature.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tenacity) })
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(tenacity) {
		t.Fatal("it returns to the battlefield")
	}
	types := effectstest.EffectiveTypes(t, g, tenacity)
	if effectstest.ContainsString(types, "Creature") || !effectstest.ContainsString(types, "Enchantment") {
		t.Errorf("it's an enchantment, not a creature: %v", types)
	}

	// Its drain still works as an enchantment.
	before = opp.Life
	effectstest.E2ChangeLife(g, me.ID, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("the returned enchantment still drains: %d", before-opp.Life)
	}

	// Dying again as a noncreature, it stays in the graveyard.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tenacity) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tenacity) {
		t.Error("a second death, as an enchantment, does not return it again")
	}
	if !me.Graveyard.Contains(tenacity) {
		t.Error("it stays in the graveyard")
	}
}

// --- K'rrik, Son of Yawgmoth --------------------------------------

func TestE2KrrikGrowsOnBlackSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	krrik := effectstest.B12Push(g, me.ID, "K'rrik, Son of Yawgmoth", "Legendary Creature — Phyrexian Horror Minion", effectstest.E2KrrikOracle, 2, 2)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, krrik), "lifelink") {
		t.Error("printed lifelink")
	}

	effectstest.E2CastSpellOfColor(t, g, "Black Thing", "Sorcery", "B")
	effectstest.PassPriorityAroundTable(t, g)
	if got := e2CountersOn(g, krrik); got != 1 {
		t.Fatalf("a black spell puts a +1/+1 counter on K'rrik: %d", got)
	}

	effectstest.E2CastSpellOfColor(t, g, "White Thing", "Sorcery", "W")
	effectstest.PassPriorityAroundTable(t, g)
	if got := e2CountersOn(g, krrik); got != 1 {
		t.Errorf("a white spell puts none: %d", got)
	}

	effectstest.E2CastSpellOfColor(t, g, "Golgari Thing", "Sorcery", "B", "G")
	effectstest.PassPriorityAroundTable(t, g)
	if got := e2CountersOn(g, krrik); got != 2 {
		t.Errorf("a black-green spell is a black spell: %d", got)
	}
}
