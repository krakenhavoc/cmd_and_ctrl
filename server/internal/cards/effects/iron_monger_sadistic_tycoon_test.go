package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// iron_monger_sadistic_tycoon_test.go — game.EventConnived (CR 701.50f,
// #4641) and the card that watches it.

const ironMongerOracle = "7324d8c0-3a30-452b-b05d-ec7b6a98666c"

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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b := ironMongerBoard{g: g, me: me, opp: opp}
	b.monger = b12Push(g, me.ID, "Iron Monger, Sadistic Tycoon", "Legendary Artifact Creature — Human Villain", ironMongerOracle, 2, 2)
	b.villain = b12Creature(g, me.ID, "My Villain", "Creature — Human Villain", 2, 2)
	b.bear = b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b.theirVillain = b12Creature(g, opp.ID, "Their Villain", "Creature — Human Villain", 2, 2)
	b.theirBear = b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
	return b
}

// connive has `controller` make `creatures` connive, as an effect does.
func (b ironMongerBoard) connive(t *testing.T, controller uuid.UUID, creatures ...uuid.UUID) {
	t.Helper()
	var err error
	b.g.WithWriteLock(func() {
		err = ConniveEach{Controller: controller, Creatures: creatures}.Apply(NewContext(b.g, &game.StackItem{
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
		if got := counterOn(b.g, c.id, game.CounterPlusOne); got != c.want {
			t.Errorf("%s has %d +1/+1 counters, want %d", c.name, got, c.want)
		}
	}
}

func TestIronMongerIsRegisteredWhole(t *testing.T) {
	spec, ok := Lookup(ironMongerOracle)
	if !ok {
		t.Fatal("Iron Monger is not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %s, caveats %v", spec.Completeness, spec.Caveats)
	}
	b := newIronMongerBoard(t)
	if !hasEffectiveKeyword(t, b.g, b.monger, "flying") {
		t.Error("Iron Monger has no flying")
	}
}

// Your creature connives: each Villain you control gets a counter, Iron
// Monger included, and nothing else does. The Bear discards a land, so its
// own connive counter does not muddy the count.
func TestIronMongerCountersEachVillainYouControlWhenYourCreatureConnives(t *testing.T) {
	b := newIronMongerBoard(t)
	land := handCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	b.connive(t, b.me.ID, b.bear)
	answerDiscard(t, b.g, b.me.ID, land)
	passPriorityAroundTable(t, b.g)
	b.counters(t, 1, 1, 0, 0, 0)
}

// The trigger is the connive, not the counter: the discard here is
// whatever the draw brought, and a conniver that gets no counter (a land
// discard above) still connived (CR 701.50f).
func TestIronMongerTriggersOnTheConniveNotTheCounter(t *testing.T) {
	b := newIronMongerBoard(t)
	b.connive(t, b.me.ID, b.bear)
	answerDiscard(t, b.g, b.me.ID, b.me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, b.g)
	if got := len(eventsOfKind(b.g, game.EventConnived)); got != 1 {
		t.Fatalf("%d connive events, want 1", got)
	}
	if got := counterOn(b.g, b.monger, game.CounterPlusOne); got != 1 {
		t.Errorf("Iron Monger has %d counters, want 1", got)
	}
}

// Two creatures connive at once (CR 701.50c: one after the other): the
// trigger is once per creature, so every Villain you control gets two.
func TestIronMongerTriggersOncePerConnivingCreature(t *testing.T) {
	b := newIronMongerBoard(t)
	land1 := handCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	land2 := handCardFull(b.me, "Island", "Basic Land — Island", "", "", nil)
	b.connive(t, b.me.ID, b.bear, b.villain)
	answerDiscard(t, b.g, b.me.ID, land1)
	answerDiscard(t, b.g, b.me.ID, land2)
	passPriorityAroundTable(t, b.g)
	evs := eventsOfKind(b.g, game.EventConnived)
	if len(evs) != 2 || evs[0].CardID != b.bear || evs[1].CardID != b.villain {
		t.Fatalf("connive events %+v, want the Bear then the Villain", evs)
	}
	b.counters(t, 2, 2, 0, 0, 0)
}

// An opponent's creature connives: not yours, no trigger.
func TestIronMongerIgnoresAnOpponentsCreatureConniving(t *testing.T) {
	b := newIronMongerBoard(t)
	land := handCardFull(b.opp, "Forest", "Basic Land — Forest", "", "", nil)
	b.connive(t, b.opp.ID, b.theirBear)
	answerDiscard(t, b.g, b.opp.ID, land)
	passPriorityAroundTable(t, b.g)
	if got := len(eventsOfKind(b.g, game.EventConnived)); got != 1 {
		t.Fatalf("%d connive events, want 1 (the opponent's still happened)", got)
	}
	b.counters(t, 0, 0, 0, 0, 0)
}

// CR 701.50b: a creature that left the battlefield as it was told to
// connive still connives, as its last controller's: the draw and discard
// happen, there is no counter, and the event and the trigger still come.
func TestIronMongerTriggersForACreatureThatLeftBeforeItConnived(t *testing.T) {
	b := newIronMongerBoard(t)
	land := handCardFull(b.me, "Forest", "Basic Land — Forest", "", "", nil)
	moveToGraveyard(t, b.g, b.me.ID, b.bear)
	b.connive(t, b.me.ID, b.bear)
	answerDiscard(t, b.g, b.me.ID, land)
	passPriorityAroundTable(t, b.g)
	evs := eventsOfKind(b.g, game.EventConnived)
	if len(evs) != 1 || evs[0].CardID != b.bear || evs[0].Actor != b.me.ID {
		t.Fatalf("connive events %+v, want one for the departed Bear under me", evs)
	}
	b.counters(t, 1, 1, -1, 0, 0)
}

// A nonland discard grows the conniver itself (CR 701.50a); a Villain
// conniver then gets Iron Monger's counter on top.
func TestIronMongerCounterStacksOnTheConniversOwn(t *testing.T) {
	b := newIronMongerBoard(t)
	spell := handCardFull(b.me, "Spell A", "Sorcery", "{1}", "", nil)
	b.connive(t, b.me.ID, b.villain)
	answerDiscard(t, b.g, b.me.ID, spell)
	passPriorityAroundTable(t, b.g)
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
	answerDiscard(t, b.g, b.me.ID, ids...)
	passPriorityAroundTable(t, b.g)
	if got := len(eventsOfKind(b.g, game.EventConnived)); got != 1 {
		t.Fatalf("%d connive events, want 1", got)
	}
	if got := counterOn(b.g, b.monger, game.CounterPlusOne); got != 1 {
		t.Errorf("Iron Monger has %d counters, want 1", got)
	}
}

// Connive 0 is no connive (CR 701.50e): no event, no trigger.
func TestConniveZeroEmitsNothing(t *testing.T) {
	b := newIronMongerBoard(t)
	b.conniveX(t, 0)
	if got := len(eventsOfKind(b.g, game.EventConnived)); got != 0 {
		t.Errorf("%d connive events for connive 0, want 0", got)
	}
}

// conniveX has my Bear connive X.
func (b ironMongerBoard) conniveX(t *testing.T, x int) {
	t.Helper()
	var err error
	b.g.WithWriteLock(func() {
		err = ConniveX{Controller: b.me.ID, Creature: b.bear, X: x}.Apply(NewContext(b.g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: b.me.ID, SourceCardID: b.bear,
		}))
	})
	if err != nil {
		t.Fatalf("ConniveX: %v", err)
	}
}
