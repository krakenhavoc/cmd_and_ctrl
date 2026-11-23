package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ironMongerBoard is a table where seat 0 controls Iron Monger, another
// Villain and a Bear, and seat 1 a Villain and a Bear. Both hands are
// emptied so a test puts known cards in them.
type ironMongerBoard struct {
	g                       *game.Game
	me, opp                 *game.Player
	monger, villain, bear   uuid.UUID
	theirVillain, theirBear uuid.UUID
}

func newIronMongerBoard(t *testing.T) ironMongerBoard {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b := ironMongerBoard{g: g, me: me, opp: opp}
	b.monger = effectstest.B12Push(g, me.ID, "Iron Monger, Sadistic Tycoon", "Legendary Artifact Creature — Human Villain", effectstest.IronMongerOracle, 2, 2)
	b.villain = effectstest.B12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	b.bear = effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b.theirVillain = effectstest.B12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	b.theirBear = effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
	return b
}

// connive has `controller` make `creatures` connive, as an effect does.
func (b ironMongerBoard) connive(t *testing.T, controller uuid.UUID, creatures ...uuid.UUID) {
	t.Helper()
	var err error
	b.g.WithWriteLock(func() {
		err = effects.ConniveEach{Controller: controller, Creatures: creatures}.Apply(effects.NewContext(b.g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: controller, SourceCardID: creatures[0],
		}))
	})
	if err != nil {
		t.Fatalf("ConniveEach: %v", err)
	}
}

// counters checks the +1/+1 counters on all five creatures.
func (b ironMongerBoard) counters(t *testing.T, monger, villain, bear, theirVillain, theirBear int) {
	t.Helper()
	for _, c := range []struct {
		name string
		id   uuid.UUID
		want int
	}{
		{"Iron Monger", b.monger, monger}, {"my Villain", b.villain, villain}, {"my Bear", b.bear, bear},
		{"their Villain", b.theirVillain, theirVillain}, {"their Bear", b.theirBear, theirBear},
	} {
		if got := effectstest.CounterOn(b.g, c.id, game.CounterPlusOne); got != c.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", c.name, got, c.want)
		}
	}
}

func TestIronMongerIsRegisteredWhole(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.IronMongerOracle)
	if !ok {
		t.Fatal("Iron Monger is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	b := newIronMongerBoard(t)
	if !effectstest.HasEffectiveKeyword(t, b.g, b.monger, "flying") {
		t.Error("Iron Monger has no flying")
	}
}

// Your creature connives: each Villain you control gets a counter, Iron
// Monger included, and nothing else does. The Bear discards a land, so its
// own connive counter does not muddy the count.
func TestIronMongerCountersEachVillainYouControlWhenYourCreatureConnives(t *testing.T) {
	b := newIronMongerBoard(t)
	land := effectstest.HandCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	b.connive(t, b.me.ID, b.bear)
	effectstest.AnswerDiscard(t, b.g, b.me.ID, land)
	effectstest.PassPriorityAroundTable(t, b.g)
	b.counters(t, 1, 1, 0, 0, 0)
}

// The trigger is the connive, not the counter: the discard here is
// whatever the draw brought, and a conniver that gets no counter (a land
// discard above) still connived (CR 701.50f).
func TestIronMongerTriggersOnTheConniveNotTheCounter(t *testing.T) {
	b := newIronMongerBoard(t)
	b.connive(t, b.me.ID, b.bear)
	effectstest.AnswerDiscard(t, b.g, b.me.ID, b.me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, b.g)
	if got := len(effectstest.EventsOfKind(b.g, game.EventConnived)); got != 1 {
		t.Fatalf("%d connive events, want 1", got)
	}
	if got := effectstest.CounterOn(b.g, b.monger, game.CounterPlusOne); got != 1 {
		t.Errorf("Iron Monger has %d counters, want 1", got)
	}
}

// Two creatures connive at once (CR 701.50c: one after the other): the
// trigger is once per creature, so every Villain you control gets two.
func TestIronMongerTriggersOncePerConnivingCreature(t *testing.T) {
	b := newIronMongerBoard(t)
	land1 := effectstest.HandCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	land2 := effectstest.HandCardFull(b.me, "Island", "Basic Land — Island", "", "", nil)
	b.connive(t, b.me.ID, b.bear, b.villain)
	effectstest.AnswerDiscard(t, b.g, b.me.ID, land1)
	effectstest.AnswerDiscard(t, b.g, b.me.ID, land2)
	effectstest.PassPriorityAroundTable(t, b.g)
	evs := effectstest.EventsOfKind(b.g, game.EventConnived)
	if len(evs) != 2 || evs[0].CardID != b.bear || evs[1].CardID != b.villain {
		t.Fatalf("connive events %+v, want the Bear then the Villain", evs)
	}
	b.counters(t, 2, 2, 0, 0, 0)
}

// An opponent's creature connives: not yours, no trigger.
func TestIronMongerIgnoresAnOpponentsCreatureConniving(t *testing.T) {
	b := newIronMongerBoard(t)
	land := effectstest.HandCardFull(b.opp, "Forest", "Basic Land — Forest", "", "", nil)
	b.connive(t, b.opp.ID, b.theirBear)
	effectstest.AnswerDiscard(t, b.g, b.opp.ID, land)
	effectstest.PassPriorityAroundTable(t, b.g)
	if got := len(effectstest.EventsOfKind(b.g, game.EventConnived)); got != 1 {
		t.Fatalf("%d connive events, want 1 (the opponent's still happened)", got)
	}
	b.counters(t, 0, 0, 0, 0, 0)
}

// CR 701.50b: a creature that left the battlefield as it was told to
// connive still connives, as its last controller's: the draw and discard
// happen, there is no counter, and the event and the trigger still come.
func TestIronMongerTriggersForACreatureThatLeftBeforeItConnived(t *testing.T) {
	b := newIronMongerBoard(t)
	land := effectstest.HandCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	effectstest.MoveToGraveyard(t, b.g, b.me.ID, b.bear)
	b.connive(t, b.me.ID, b.bear)
	effectstest.AnswerDiscard(t, b.g, b.me.ID, land)
	effectstest.PassPriorityAroundTable(t, b.g)
	evs := effectstest.EventsOfKind(b.g, game.EventConnived)
	if len(evs) != 1 || evs[0].CardID != b.bear || evs[0].Actor != b.me.ID {
		t.Fatalf("connive events %+v, want one for the departed Bear under me", evs)
	}
	b.counters(t, 1, 1, -1, 0, 0)
}

// A nonland discard grows the conniver itself (CR 701.50a); a Villain
// conniver then gets Iron Monger's counter on top.
func TestIronMongerCounterStacksOnTheConniversOwn(t *testing.T) {
	b := newIronMongerBoard(t)
	spell := effectstest.HandCardFull(b.me, "Spell A", "Sorcery", "{1}", "", nil)
	b.connive(t, b.me.ID, b.villain)
	effectstest.AnswerDiscard(t, b.g, b.me.ID, spell)
	effectstest.PassPriorityAroundTable(t, b.g)
	b.counters(t, 1, 2, 0, 0, 0)
}

// ConniveX is one connive however many cards: one event, one trigger.
func TestConniveXEmitsOneEvent(t *testing.T) {
	b := newIronMongerBoard(t)
	b.conniveX(t, 2)
	var ids []uuid.UUID
	for _, c := range b.me.Hand.Cards {
		ids = append(ids, c.InstanceID)
	}
	effectstest.AnswerDiscard(t, b.g, b.me.ID, ids...)
	effectstest.PassPriorityAroundTable(t, b.g)
	if got := len(effectstest.EventsOfKind(b.g, game.EventConnived)); got != 1 {
		t.Fatalf("%d connive events, want 1", got)
	}
	if got := effectstest.CounterOn(b.g, b.monger, game.CounterPlusOne); got != 1 {
		t.Errorf("Iron Monger has %d counters, want 1", got)
	}
}

// Connive 0 is no connive (CR 701.50e): no event, no trigger.
func TestConniveZeroEmitsNothing(t *testing.T) {
	b := newIronMongerBoard(t)
	b.conniveX(t, 0)
	if got := len(effectstest.EventsOfKind(b.g, game.EventConnived)); got != 0 {
		t.Errorf("%d connive events for connive 0, want 0", got)
	}
}

// conniveX has my Bear connive X.
func (b ironMongerBoard) conniveX(t *testing.T, x int) {
	t.Helper()
	var err error
	b.g.WithWriteLock(func() {
		err = effects.ConniveX{Controller: b.me.ID, Creature: b.bear, X: x}.Apply(effects.NewContext(b.g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: b.me.ID, SourceCardID: b.bear,
		}))
	})
	if err != nil {
		t.Fatalf("ConniveX: %v", err)
	}
}
