package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// issue1117_commander_test.go — the #1117 Betor life-swap deck's
// commander and its three lifegain payoffs.

const (
	e2BetorOracle            = "990b5e12-6e04-4832-9d64-87278f12cbda"
	e2RodolfOracle           = "26a716b2-a2e2-405b-accd-df572416d8bf"
	e2EnduringTenacityOracle = "98e698ae-1a69-469c-9cfb-0e3fedeb71d4"
	e2KrrikOracle            = "cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd"
)

func TestE2CommanderBatchRegisters(t *testing.T) {
	want := map[string]string{
		e2BetorOracle:            "Betor, Ancestor's Voice",
		e2RodolfOracle:           "Rodolf Duskbringer",
		e2EnduringTenacityOracle: "Enduring Tenacity",
		e2KrrikOracle:            "K'rrik, Son of Yawgmoth",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s registered as %q", oracle, spec.Name)
		}
	}
}

// e2ChangeLife is the drain / gain the tally counts.
func e2ChangeLife(g *game.Game, player uuid.UUID, delta int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, player, delta) })
}

// e2CastSpellOfColor casts a non-catalog spell whose only interesting
// property is its colour — K'rrik's trigger reads the spell's COLOUR,
// not a pip in its cost, so the colour is set directly and the cost
// left empty.
func e2CastSpellOfColor(t *testing.T, g *game.Game, name, typeLine string, colors ...string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Colors: colors,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
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
// one", and they arrive as two separate triggers (the declared
// caveat), so the prompts are answered by what each one offers rather
// than by a fixed order.
func TestE2BetorCountersAndReanimatesOffTheTurnsLifeSwing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	betor := b12Push(g, me.ID, "Betor, Ancestor's Voice", "Legendary Creature — Spirit Dragon", e2BetorOracle, 3, 5)
	abilities := effectiveAbilities(t, g, betor)
	if !hasAbility(abilities, "flying") || !hasAbility(abilities, "lifelink") {
		t.Errorf("printed flying and lifelink: %v", abilities)
	}
	ally := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	cheap := pushGraveyardPermanent(me, "Dead Scout", "Creature — Human Scout", "{B}")
	expensive := pushGraveyardPermanent(me, "Dead Fatty", "Creature — Giant", "{5}{B}")
	notMine := pushGraveyardPermanent(opp, "Their Dead Bear", "Creature — Bear", "{B}")

	// Three gained, two lost: three counters, and a reanimation
	// budget of two mana value.
	e2ChangeLife(g, me.ID, 3)
	e2ChangeLife(g, me.ID, -2)
	advanceToEndStepOf(t, g, 0)

	sawCounters, sawReturn := false, false
	for i := 0; i < 6; i++ {
		b04WaitForPick(t, g, me.ID)
		p := latestPickTarget(g, me.ID)
		switch {
		case hasID(p.PickTargetCards, ally):
			sawCounters = true
			if p.PickTargetMin != 0 {
				t.Errorf("up to ONE other target creature you control: min %d", p.PickTargetMin)
			}
			if hasID(p.PickTargetCards, betor) {
				t.Error("OTHER target creature — Betor may not choose itself")
			}
			if hasID(p.PickTargetCards, theirs) {
				t.Error("a creature you control — not an opponent's")
			}
			pickCard(t, g, me.ID, ally)
		case hasID(p.PickTargetCards, cheap):
			sawReturn = true
			if hasID(p.PickTargetCards, expensive) {
				t.Error("mana value 6 is above the two life lost this turn")
			}
			if hasID(p.PickTargetCards, notMine) {
				t.Error("from YOUR graveyard only")
			}
			pickCard(t, g, me.ID, cheap)
		default:
			t.Fatalf("unexpected pick_target prompt: %+v", p)
		}
		if sawCounters && sawReturn {
			break
		}
	}
	if !sawCounters || !sawReturn {
		t.Fatalf("both end-step clauses ask for a target (counters=%v return=%v)", sawCounters, sawReturn)
	}
	// #1529: the two targeted end-step triggers are one CR 603.3b
	// batch, ordered once both targets are chosen.
	if !answerTriggerOrderLastQueuedFirst(t, g) {
		t.Error("Betor's two end-step triggers were not offered for ordering")
	}
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Betor, Ancestor's Voice", "Legendary Creature — Spirit Dragon", e2BetorOracle, 3, 5)
	ally := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	dead := pushGraveyardPermanent(me, "Dead Scout", "Creature — Human Scout", "{B}")

	advanceToEndStepOf(t, g, 0)
	// The counters clause still targets — it is not conditional — so
	// answer it; the graveyard clause has no legal target at a budget
	// of zero and never asks.
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetCards, dead) {
		t.Fatal("nothing was lost this turn: no creature card is within the budget")
	}
	pickCard(t, g, me.ID, ally)
	passPriorityAroundTable(t, g)

	if got := e2CountersOn(g, ally); got != 0 {
		t.Errorf("no life gained, no counters: %d", got)
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("no life lost, nothing reanimated")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("no prompt is left waiting")
	}
}

// --- Rodolf Duskbringer -------------------------------------------

func TestE2RodolfGainsIndestructibleThenPaysToReanimate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rodolf := b12Push(g, me.ID, "Rodolf Duskbringer", "Legendary Creature — Vampire Angel", e2RodolfOracle, 4, 4)
	abilities := effectiveAbilities(t, g, rodolf)
	for _, kw := range []string{"flying", "deathtouch", "lifelink"} {
		if !hasAbility(abilities, kw) {
			t.Errorf("printed %s: %v", kw, abilities)
		}
	}
	if hasAbility(abilities, "indestructible") {
		t.Error("indestructible is granted by the lifegain trigger, not printed")
	}
	cheap := pushGraveyardPermanent(me, "Dead Scout", "Creature — Human Scout", "{1}{B}")
	expensive := pushGraveyardPermanent(me, "Dead Fatty", "Creature — Giant", "{5}{B}")

	e2ChangeLife(g, me.ID, 5)
	passPriorityAroundTable(t, g)
	if !hasAbility(effectiveAbilities(t, g, rodolf), "indestructible") {
		t.Fatal("gaining life grants indestructible until end of turn")
	}

	advanceToEndStepOf(t, g, 0)
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the end step offers the optional {1}{W/B}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	answerPayUnless(t, g, me.ID, true)

	// "When you do" is a reflexive trigger with its own target, asked
	// after the payment.
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if !hasID(p.PickTargetCards, cheap) {
		t.Fatalf("mana value 2 is within the 5 life gained: %+v", p)
	}
	if hasID(p.PickTargetCards, expensive) {
		t.Error("mana value 6 is above the 5 life gained")
	}
	if p.PickTargetMin != 1 {
		t.Errorf("a required target, not an up-to-one: min %d", p.PickTargetMin)
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(cheap) {
		t.Error("paying returns the chosen creature card to the battlefield")
	}
	if !me.Graveyard.Contains(expensive) {
		t.Error("only the chosen card comes back")
	}
}

// Declining the optional cost raises no reflexive trigger at all.
func TestE2RodolfDecliningThePaymentReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Rodolf Duskbringer", "Legendary Creature — Vampire Angel", e2RodolfOracle, 4, 4)
	dead := pushGraveyardPermanent(me, "Dead Scout", "Creature — Human Scout", "{1}{B}")

	e2ChangeLife(g, me.ID, 5)
	passPriorityAroundTable(t, g)
	advanceToEndStepOf(t, g, 0)
	for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)

	if latestPickTarget(g, me.ID) != nil {
		t.Error("declining asks for no target")
	}
	if !me.Graveyard.Contains(dead) {
		t.Error("declining returns nothing")
	}
}

// --- Enduring Tenacity --------------------------------------------

func TestE2EnduringTenacityDrainsAndStaysDead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tenacity := b12Push(g, me.ID, "Enduring Tenacity", "Enchantment Creature — Snake Glimmer", e2EnduringTenacityOracle, 4, 3)
	before := opp.Life

	e2ChangeLife(g, me.ID, 4)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if hasID(p.PickTargetPlayers, me.ID) {
		t.Error("target OPPONENT — not yourself")
	}
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != before-4 {
		t.Errorf("the opponent loses the 4 life gained: %d", before-opp.Life)
	}

	// The declared caveat: it dies like any other creature.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(tenacity) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tenacity) {
		t.Error("it does not return to the battlefield as an enchantment")
	}
	if !me.Graveyard.Contains(tenacity) {
		t.Error("it goes to the graveyard")
	}
}

// --- K'rrik, Son of Yawgmoth --------------------------------------

func TestE2KrrikGrowsOnBlackSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	krrik := b12Push(g, me.ID, "K'rrik, Son of Yawgmoth", "Legendary Creature — Phyrexian Horror Minion", e2KrrikOracle, 2, 2)
	if !hasAbility(effectiveAbilities(t, g, krrik), "lifelink") {
		t.Error("printed lifelink")
	}

	e2CastSpellOfColor(t, g, "Black Thing", "Sorcery", "B")
	passPriorityAroundTable(t, g)
	if got := e2CountersOn(g, krrik); got != 1 {
		t.Fatalf("a black spell puts a +1/+1 counter on K'rrik: %d", got)
	}

	e2CastSpellOfColor(t, g, "White Thing", "Sorcery", "W")
	passPriorityAroundTable(t, g)
	if got := e2CountersOn(g, krrik); got != 1 {
		t.Errorf("a white spell puts none: %d", got)
	}

	e2CastSpellOfColor(t, g, "Golgari Thing", "Sorcery", "B", "G")
	passPriorityAroundTable(t, g)
	if got := e2CountersOn(g, krrik); got != 2 {
		t.Errorf("a black-green spell is a black spell: %d", got)
	}
}
