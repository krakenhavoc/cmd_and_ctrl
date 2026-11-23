package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// miTrimHands cuts every seat's hand down to one card so a test that
// walks several turns is never stopped by a cleanup discard.
func miTrimHands(g *game.Game) {
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			if p.Hand.Size() > 1 {
				p.Hand.Cards = p.Hand.Cards[:1]
			}
		}
	})
}

// miCastOverlord seeds Overlord of the Balemurk in the active seat's
// hand WITH its printed 5/5 and casts it, for its impending cost when
// altCost is set and for its mana cost when it is empty.
//
// castWithAltCost is not enough on its own: it seeds no power or
// toughness (printed stats reach a real card through deck import),
// and a woken Overlord that is a 0/0 is killed by the next
// state-based check for reasons that have nothing to do with the
// card under test.
func miCastOverlord(t *testing.T, g *game.Game, altCost string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Overlord of the Balemurk",
		TypeLine:   "Enchantment Creature — Avatar Horror",
		OracleID:   effectstest.BalemurkOracle,
		ManaCost:   "{3}{B}{B}",
		Power:      5,
		Toughness:  5,
		Owner:      active.ID,
		Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{AlternativeCost: altCost}); err != nil {
		t.Fatalf("CastSpell Overlord of the Balemurk (%q): %v", altCost, err)
	}
	return id
}

// --- Barrowgoyf ---------------------------------------------------

// The CDA, Tarmogoyf's: power is the number of distinct card types
// across EVERY graveyard at the table, toughness that plus one. Two
// instants in one pile are one type; an opponent's graveyard counts.
func TestBarrowgoyfSizeIsCardTypesInAllGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goyf := effectstest.B12Push(g, me.ID, "Barrowgoyf", "Creature — Lhurgoyf", effectstest.BarrowgoyfOracle, 0, 1)

	if got := effectstest.EffectivePower(t, g, goyf); got != 0 {
		t.Errorf("empty graveyards: power %d, want 0", got)
	}
	if got := effectstest.EffectiveToughness(t, g, goyf); got != 1 {
		t.Errorf("empty graveyards: toughness %d, want 1", got)
	}

	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Grizzly Bears", "Creature — Bear")
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[2].ID, "Lightning Bolt", "Instant")
	// A second instant, in a third player's graveyard, is still one
	// type — the clause counts TYPES, not cards, across one union.
	effectstest.PushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Giant Growth", "Instant")

	if got := effectstest.EffectivePower(t, g, goyf); got != 2 {
		t.Errorf("creature + instant: power %d, want 2", got)
	}
	if got := effectstest.EffectiveToughness(t, g, goyf); got != 3 {
		t.Errorf("creature + instant: toughness %d, want 3", got)
	}

	abilities := effectstest.EffectiveAbilities(t, g, goyf)
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !effectstest.HasString(abilities, kw) {
			t.Errorf("Barrowgoyf is missing %q: %v", kw, abilities)
		}
	}
}

// The size follows a mill and a discard at once, with nothing else on
// the battlefield moving: a card reaching (or leaving) a graveyard
// refreshes the layer cache on its own (game/layer_listener.go).
func TestBarrowgoyfSizeUpdatesAtOnceAfterAMillOrADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goyf := effectstest.B12Push(g, me.ID, "Barrowgoyf", "Creature — Lhurgoyf", effectstest.BarrowgoyfOracle, 0, 1)
	if got := effectstest.EffectivePower(t, g, goyf); got != 0 {
		t.Fatalf("empty graveyards: power %d, want 0", got)
	}

	effectstest.PlTop(me, "Mind Stone", "Artifact", "{2}")
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(me.ID, 1, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	if got, want := effectstest.EffectivePower(t, g, goyf), 1; got != want {
		t.Errorf("after milling an artifact: power %d, want %d", got, want)
	}
	if got, want := effectstest.EffectiveToughness(t, g, goyf), 2; got != want {
		t.Errorf("after milling an artifact: toughness %d, want %d", got, want)
	}

	sorcery := uuid.New()
	me.Hand.Cards = nil
	me.Hand.PushTop(game.Card{InstanceID: sorcery, Name: "Divination", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 1); err != nil {
			t.Fatalf("discard: %v", err)
		}
	})
	if !me.Graveyard.Contains(sorcery) {
		t.Fatal("the sorcery was not discarded")
	}
	if got, want := effectstest.EffectivePower(t, g, goyf), 2; got != want {
		t.Errorf("after discarding a sorcery: power %d, want %d", got, want)
	}
}

// The combat trigger, end to end: two damage means "you may mill
// two", the mill is a decision asked during the resolution, and the
// creature card among the two milled is offered back to hand.
func TestBarrowgoyfCombatDamageMillsAndTakesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Two card types in the graveyards makes Barrowgoyf a 2/3, so
	// "that many" is two and the test knows what to expect.
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Grizzly Bears", "Creature — Bear")
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Lightning Bolt", "Instant")
	goyf := effectstest.B12Push(g, me.ID, "Barrowgoyf", "Creature — Lhurgoyf", effectstest.BarrowgoyfOracle, 0, 1)

	land := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")
	beast := effectstest.PlTop(me, "Rampant Beast", "Creature — Beast", "{3}{G}")

	effectstest.AttackWith(t, g, opp.ID, goyf)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40-2 {
		t.Fatalf("defender life %d, want 38 (an unblocked 2/3)", opp.Life)
	}

	graveBefore := me.Graveyard.Size()
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if got := me.Graveyard.Size() - graveBefore; got != 2 {
		t.Fatalf("milled %d cards, want 2", got)
	}

	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature was offered from among the milled cards")
	}
	if effectstest.HasID(pick.ChooseCards, land) {
		t.Error("the land was offered; only a creature card may be taken")
	}
	if !effectstest.HasID(pick.ChooseCards, beast) {
		t.Fatalf("the milled creature was not offered: %v", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, beast)
	if !effectstest.MiInHand(me, beast) {
		t.Error("the chosen creature did not reach hand")
	}
}

// Declining the mill is a real answer, and it stops there — no
// second prompt, nothing milled.
func TestBarrowgoyfDeclinedMillDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushGraveyardCardWithTypeLine(g, me.ID, "Grizzly Bears", "Creature — Bear")
	goyf := effectstest.B12Push(g, me.ID, "Barrowgoyf", "Creature — Lhurgoyf", effectstest.BarrowgoyfOracle, 0, 1)
	effectstest.PlTop(me, "Rampant Beast", "Creature — Beast", "{3}{G}")

	effectstest.AttackWith(t, g, opp.ID, goyf)
	effectstest.PassPriorityAroundTable(t, g)
	graveBefore := me.Graveyard.Size()
	effectstest.AnswerMayChoice(t, g, me.ID, false)

	if got := me.Graveyard.Size(); got != graveBefore {
		t.Errorf("graveyard grew to %d from %d on a declined mill", got, graveBefore)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the take prompt was offered although nothing was milled")
	}
}

// --- Six ------------------------------------------------------------

// The attack trigger: three cards off the top, mandatory, and a land
// from among them offered back.
func TestSixAttackMillsThreeAndOffersALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	six := effectstest.B12Push(g, me.ID, "Six", "Legendary Creature — Treefolk", effectstest.SixOracle, 2, 4)

	bolt := effectstest.PlTop(me, "Lightning Bolt", "Instant", "{R}")
	bear := effectstest.PlTop(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	forest := effectstest.PlTop(me, "Forest", "Basic Land — Forest", "")

	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, six), "reach") {
		t.Error("Six is missing reach")
	}

	graveBefore := me.Graveyard.Size()
	effectstest.DeclareAttack(t, g, opp.ID, six)
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Graveyard.Size() - graveBefore; got != 3 {
		t.Fatalf("milled %d cards, want 3", got)
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land was offered from among the milled cards")
	}
	if effectstest.HasID(pick.ChooseCards, bolt) || effectstest.HasID(pick.ChooseCards, bear) {
		t.Errorf("a nonland card was offered: %v", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, forest)
	if !effectstest.MiInHand(me, forest) {
		t.Error("the chosen land did not reach hand")
	}
}

// --- Impending, and Overlord of the Balemurk ------------------------

// The keyword is wired under the key the client sends, at the price
// and counter count the card prints.
func TestImpendingIsWiredForOverlordOfTheBalemurk(t *testing.T) {
	ac := game.AlternativeCostByKey(effectstest.BalemurkOracle, effects.AltCostKeyImpending)
	if ac == nil {
		t.Fatal("Overlord of the Balemurk offers no impending cost")
	}
	if ac.ManaCost != "{1}{B}" {
		t.Errorf("impending cost %q, want {1}{B}", ac.ManaCost)
	}
	if ac.EntersWithCounterName != game.CounterTime || ac.EntersWithCounterCount != 5 {
		t.Errorf("impending counters %d %q, want 5 time",
			ac.EntersWithCounterCount, ac.EntersWithCounterName)
	}
	if ac.Label != "Impending 5—{1}{B}" {
		t.Errorf("impending label %q", ac.Label)
	}
	// Impending opens no cast zone and exiles nothing: an Overlord
	// that dies is an ordinary card in an ordinary graveyard.
	if ac.FromZone != "" || ac.ExileOnLeavingStack {
		t.Errorf("impending bound a zone (%q) or an exile (%v) it must not",
			ac.FromZone, ac.ExileOnLeavingStack)
	}
}

// Cast for the impending cost: five time counters, not a creature,
// and the enters half of the trigger still fires.
func TestOverlordImpendedEntersWithTimeCountersAndIsNoCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	graveBefore := me.Graveyard.Size()

	id := miCastOverlord(t, g, effects.AltCostKeyImpending)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("the Overlord did not enter the battlefield")
	}
	if got := effectstest.MiCounters(t, g, id, game.CounterTime); got != 5 {
		t.Errorf("time counters %d, want 5", got)
	}
	types := effectstest.EffectiveTypes(t, g, id)
	if effectstest.HasString(types, "Creature") {
		t.Errorf("an impending Overlord is still a creature: %v", types)
	}
	if !effectstest.HasString(types, "Enchantment") {
		t.Errorf("the Enchantment half was dropped too: %v", types)
	}
	if got := me.Graveyard.Size() - graveBefore; got != 4 {
		t.Errorf("the enters trigger milled %d, want 4", got)
	}
}

// Cast for its printed cost it is an ordinary 5/5 at once: no
// counters, no countdown, and the same enters trigger.
func TestOverlordHardCastIsACreatureImmediately(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	graveBefore := me.Graveyard.Size()

	id := miCastOverlord(t, g, "")
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.MiCounters(t, g, id, game.CounterTime); got != 0 {
		t.Errorf("a hard-cast Overlord carries %d time counters, want 0", got)
	}
	if !effectstest.HasString(effectstest.EffectiveTypes(t, g, id), "Creature") {
		t.Errorf("a hard-cast Overlord is not a creature: %v", effectstest.EffectiveTypes(t, g, id))
	}
	if got := me.Graveyard.Size() - graveBefore; got != 4 {
		t.Errorf("the enters trigger milled %d, want 4", got)
	}
}

// The countdown: one counter comes off at each of the controller's
// end steps, the permanent cannot attack while any remain, and the
// last one leaving turns it into the creature it prints.
func TestOverlordCountdownWakesTheCreatureUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	id := miCastOverlord(t, g, effects.AltCostKeyImpending)
	effectstest.PassPriorityAroundTable(t, g)
	miTrimHands(g)

	for want := 4; want >= 0; want-- {
		miAdvancePastMyEndStep(t, g, 0)
		if got := effectstest.MiCounters(t, g, id, game.CounterTime); got != want {
			t.Fatalf("after end step %d: %d time counters, want %d", 5-want, got, want)
		}
		if want > 0 {
			// Still an enchantment, so it cannot be declared as an
			// attacker however long it has been on the battlefield.
			effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(id, opp.ID); err == nil {
				t.Fatal("an Overlord with time counters was allowed to attack")
			}
		}
	}

	if !effectstest.HasString(effectstest.EffectiveTypes(t, g, id), "Creature") {
		t.Errorf("the last counter came off and it is still not a creature: %v",
			effectstest.EffectiveTypes(t, g, id))
	}

	// And now it attacks, which fires the other half of the same
	// printed trigger.
	graveBefore := me.Graveyard.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("the woken Overlord could not attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - graveBefore; got != 4 {
		t.Errorf("the attacks trigger milled %d, want 4", got)
	}
}

// miAdvancePastMyEndStep walks to the named seat's next end step,
// resolves whatever the step put on the stack, and steps off it so
// the next call finds the FOLLOWING one.
func miAdvancePastMyEndStep(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	effectstest.AdvanceToEndStepOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep off the end step: %v", err)
	}
}
