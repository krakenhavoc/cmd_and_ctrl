package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_impending_test.go — #1117 batch E4: Barrowgoyf, Six and
// Overlord of the Balemurk, plus the impending keyword the Overlord
// brought with it.

const (
	barrowgoyfOracle = "74c0164c-130f-4572-9066-626c77f6e2ff"
	sixOracle        = "dbcbdf37-c40f-4068-b4a7-a849cab1056c"
	balemurkOracle   = "652e87af-7cf0-407f-9fb7-1a630fb8dd47"
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

// miCounters reads a battlefield card's counter pile.
func miCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s is not in the game", id)
	}
	return c.Counters[kind]
}

// miInHand reports whether the named player holds the card.
func miInHand(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
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
		OracleID:   balemurkOracle,
		ManaCost:   "{3}{B}{B}",
		Power:      5,
		Toughness:  5,
		Owner:      active.ID,
		Controller: active.ID,
	})
	advanceToMain(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	goyf := b12Push(g, me.ID, "Barrowgoyf", "Creature — Lhurgoyf", barrowgoyfOracle, 0, 1)

	if got := effectivePower(t, g, goyf); got != 0 {
		t.Errorf("empty graveyards: power %d, want 0", got)
	}
	if got := effectiveToughness(t, g, goyf); got != 1 {
		t.Errorf("empty graveyards: toughness %d, want 1", got)
	}

	pushGraveyardCardWithTypeLine(g, me.ID, "Grizzly Bears", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, g.Seats[2].ID, "Lightning Bolt", "Instant")
	// A second instant, in a third player's graveyard, is still one
	// type — the clause counts TYPES, not cards, across one union.
	pushGraveyardCardWithTypeLine(g, g.Seats[1].ID, "Giant Growth", "Instant")

	if got := effectivePower(t, g, goyf); got != 2 {
		t.Errorf("creature + instant: power %d, want 2", got)
	}
	if got := effectiveToughness(t, g, goyf); got != 3 {
		t.Errorf("creature + instant: toughness %d, want 3", got)
	}

	abilities := effectiveAbilities(t, g, goyf)
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !hasString(abilities, kw) {
			t.Errorf("Barrowgoyf is missing %q: %v", kw, abilities)
		}
	}
}

// The combat trigger, end to end: two damage means "you may mill
// two", the mill is a decision asked during the resolution, and the
// creature card among the two milled is offered back to hand.
func TestBarrowgoyfCombatDamageMillsAndTakesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Two card types in the graveyards makes Barrowgoyf a 2/3, so
	// "that many" is two and the test knows what to expect.
	pushGraveyardCardWithTypeLine(g, me.ID, "Grizzly Bears", "Creature — Bear")
	pushGraveyardCardWithTypeLine(g, me.ID, "Lightning Bolt", "Instant")
	goyf := b12Push(g, me.ID, "Barrowgoyf", "Creature — Lhurgoyf", barrowgoyfOracle, 0, 1)

	land := plTop(me, "Forest", "Basic Land — Forest", "")
	beast := plTop(me, "Rampant Beast", "Creature — Beast", "{3}{G}")

	attackWith(t, g, opp.ID, goyf)
	passPriorityAroundTable(t, g)
	if opp.Life != 40-2 {
		t.Fatalf("defender life %d, want 38 (an unblocked 2/3)", opp.Life)
	}

	graveBefore := me.Graveyard.Size()
	answerMayChoice(t, g, me.ID, true)
	if got := me.Graveyard.Size() - graveBefore; got != 2 {
		t.Fatalf("milled %d cards, want 2", got)
	}

	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature was offered from among the milled cards")
	}
	if hasID(pick.ChooseCards, land) {
		t.Error("the land was offered; only a creature card may be taken")
	}
	if !hasID(pick.ChooseCards, beast) {
		t.Fatalf("the milled creature was not offered: %v", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, beast)
	if !miInHand(me, beast) {
		t.Error("the chosen creature did not reach hand")
	}
}

// Declining the mill is a real answer, and it stops there — no
// second prompt, nothing milled.
func TestBarrowgoyfDeclinedMillDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushGraveyardCardWithTypeLine(g, me.ID, "Grizzly Bears", "Creature — Bear")
	goyf := b12Push(g, me.ID, "Barrowgoyf", "Creature — Lhurgoyf", barrowgoyfOracle, 0, 1)
	plTop(me, "Rampant Beast", "Creature — Beast", "{3}{G}")

	attackWith(t, g, opp.ID, goyf)
	passPriorityAroundTable(t, g)
	graveBefore := me.Graveyard.Size()
	answerMayChoice(t, g, me.ID, false)

	if got := me.Graveyard.Size(); got != graveBefore {
		t.Errorf("graveyard grew to %d from %d on a declined mill", got, graveBefore)
	}
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("the take prompt was offered although nothing was milled")
	}
}

// --- Six ------------------------------------------------------------

// The attack trigger: three cards off the top, mandatory, and a land
// from among them offered back.
func TestSixAttackMillsThreeAndOffersALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	six := b12Push(g, me.ID, "Six", "Legendary Creature — Treefolk", sixOracle, 2, 4)

	bolt := plTop(me, "Lightning Bolt", "Instant", "{R}")
	bear := plTop(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	forest := plTop(me, "Forest", "Basic Land — Forest", "")

	if !hasString(effectiveAbilities(t, g, six), "reach") {
		t.Error("Six is missing reach")
	}

	graveBefore := me.Graveyard.Size()
	declareAttack(t, g, opp.ID, six)
	passPriorityAroundTable(t, g)

	if got := me.Graveyard.Size() - graveBefore; got != 3 {
		t.Fatalf("milled %d cards, want 3", got)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no land was offered from among the milled cards")
	}
	if hasID(pick.ChooseCards, bolt) || hasID(pick.ChooseCards, bear) {
		t.Errorf("a nonland card was offered: %v", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, forest)
	if !miInHand(me, forest) {
		t.Error("the chosen land did not reach hand")
	}
}

// --- Impending, and Overlord of the Balemurk ------------------------

// The keyword is wired under the key the client sends, at the price
// and counter count the card prints.
func TestImpendingIsWiredForOverlordOfTheBalemurk(t *testing.T) {
	ac := game.AlternativeCostByKey(balemurkOracle, AltCostKeyImpending)
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	graveBefore := me.Graveyard.Size()

	id := miCastOverlord(t, g, AltCostKeyImpending)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(id) {
		t.Fatal("the Overlord did not enter the battlefield")
	}
	if got := miCounters(t, g, id, game.CounterTime); got != 5 {
		t.Errorf("time counters %d, want 5", got)
	}
	types := effectiveTypes(t, g, id)
	if hasString(types, "Creature") {
		t.Errorf("an impending Overlord is still a creature: %v", types)
	}
	if !hasString(types, "Enchantment") {
		t.Errorf("the Enchantment half was dropped too: %v", types)
	}
	if got := me.Graveyard.Size() - graveBefore; got != 4 {
		t.Errorf("the enters trigger milled %d, want 4", got)
	}
}

// Cast for its printed cost it is an ordinary 5/5 at once: no
// counters, no countdown, and the same enters trigger.
func TestOverlordHardCastIsACreatureImmediately(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	graveBefore := me.Graveyard.Size()

	id := miCastOverlord(t, g, "")
	passPriorityAroundTable(t, g)

	if got := miCounters(t, g, id, game.CounterTime); got != 0 {
		t.Errorf("a hard-cast Overlord carries %d time counters, want 0", got)
	}
	if !hasString(effectiveTypes(t, g, id), "Creature") {
		t.Errorf("a hard-cast Overlord is not a creature: %v", effectiveTypes(t, g, id))
	}
	if got := me.Graveyard.Size() - graveBefore; got != 4 {
		t.Errorf("the enters trigger milled %d, want 4", got)
	}
}

// The countdown: one counter comes off at each of the controller's
// end steps, the permanent cannot attack while any remain, and the
// last one leaving turns it into the creature it prints.
func TestOverlordCountdownWakesTheCreatureUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	id := miCastOverlord(t, g, AltCostKeyImpending)
	passPriorityAroundTable(t, g)
	miTrimHands(g)

	for want := 4; want >= 0; want-- {
		miAdvancePastMyEndStep(t, g, 0)
		if got := miCounters(t, g, id, game.CounterTime); got != want {
			t.Fatalf("after end step %d: %d time counters, want %d", 5-want, got, want)
		}
		if want > 0 {
			// Still an enchantment, so it cannot be declared as an
			// attacker however long it has been on the battlefield.
			advanceTo(t, g, game.StepDeclareAttackers)
			if err := g.DeclareAttacker(id, opp.ID); err == nil {
				t.Fatal("an Overlord with time counters was allowed to attack")
			}
		}
	}

	if !hasString(effectiveTypes(t, g, id), "Creature") {
		t.Errorf("the last counter came off and it is still not a creature: %v",
			effectiveTypes(t, g, id))
	}

	// And now it attacks, which fires the other half of the same
	// printed trigger.
	graveBefore := me.Graveyard.Size()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(id, opp.ID); err != nil {
		t.Fatalf("the woken Overlord could not attack: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Graveyard.Size() - graveBefore; got != 4 {
		t.Errorf("the attacks trigger milled %d, want 4", got)
	}
}

// miAdvancePastMyEndStep walks to the named seat's next end step,
// resolves whatever the step put on the stack, and steps off it so
// the next call finds the FOLLOWING one.
func miAdvancePastMyEndStep(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	advanceToEndStepOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep off the end step: %v", err)
	}
}
