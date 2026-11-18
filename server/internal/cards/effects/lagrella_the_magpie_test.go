package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const lagrellaOracle = "6a98b586-5187-48a1-bde3-f6141f3e9cb1"

const lagrellaFollowUpLabel = "Lagrella, the Magpie — put two +1/+1 counters on the returned creature"

// lagrellaPickPrompt is the enters trigger's target prompt for `chooser`.
func lagrellaPickPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// lagrellaEnters casts Lagrella and answers its enters trigger's target
// prompt with `picks`, returning Lagrella's ID and the answer's error.
func lagrellaEnters(t *testing.T, g *game.Game, picks ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := castCatalogSpell(t, g, "Lagrella, the Magpie", "Legendary Creature — Human Soldier", lagrellaOracle, nil)
	passPriorityAroundTable(t, g)
	ch := lagrellaPickPrompt(g, me.ID)
	if ch == nil {
		t.Fatal("Lagrella's enters trigger asked for no targets")
	}
	refs := make([]game.TargetRef, 0, len(picks))
	for _, p := range picks {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: p})
	}
	return id, g.ResolvePickTargets(ch.ID, me.ID, refs)
}

// lagrellaSettle passes priority until the stack is empty.
func lagrellaSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16 && len(g.StackMeta) > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	if len(g.StackMeta) > 0 {
		t.Fatal("the stack never emptied")
	}
}

// lagrellaFollowUps counts the queued follow-up triggers.
func lagrellaFollowUps(g *game.Game) int {
	n := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && !dt.Until && dt.Label == lagrellaFollowUpLabel {
			n++
		}
	}
	return n
}

// lagrellaStackHolds reports whether a stack item with `label` is waiting.
func lagrellaStackHolds(g *game.Game, label string) int {
	n := 0
	for _, it := range g.StackMeta {
		if it != nil && it.Label == label {
			n++
		}
	}
	return n
}

// lagrellaReturned is the battlefield permanent named `name`: a card the
// return brought back is a new object with a new instance ID (CR 400.7).
func lagrellaReturned(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return game.Card{}
}

func TestLagrellaIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(lagrellaOracle)
	if !ok || spec.Name != "Lagrella, the Magpie" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Lagrella is not registered Full with no caveats: %+v", spec)
	}
}

// The heart of #5877: both creatures are exiled until Lagrella leaves.
// When it does, both come back under their owners' control (CR 610.3c);
// only the one that enters under Lagrella's controller's control — my own
// creature — gets two +1/+1 counters, by a trigger that uses the stack.
// The opponent's creature gets none, and its unfired follow-up is gone
// by the next turn.
func TestLagrellaCountersOnlyOnACardReturnedUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	lagrella, err := lagrellaEnters(t, g, mine, theirs)
	if err != nil {
		t.Fatalf("picking my creature and an opponent's: %v", err)
	}
	lagrellaSettle(t, g)
	if !g.Exile.Contains(mine) || !g.Exile.Contains(theirs) {
		t.Fatal("both targets are exiled")
	}
	if lagrellaFollowUps(g) != 0 {
		t.Fatal("no follow-up is queued while the cards are held")
	}

	destroyNow(t, g, lagrella)
	g.RunStateChecksForTest()
	if g.Exile.Contains(mine) || g.Exile.Contains(theirs) {
		t.Fatal("both cards return as Lagrella leaves")
	}
	if c := lagrellaReturned(t, g, "Their Bear"); c.Controller != opp.ID {
		t.Fatalf("the opponent's card returns under its owner's control, got %v", c.Controller)
	}
	// The counters are a triggered ability: not there as it enters.
	if got := lagrellaReturned(t, g, "My Bear").Counters[game.CounterPlusOne]; got != 0 {
		t.Fatalf("the counters are not 'enters with': %d before the trigger resolved", got)
	}
	if got := lagrellaStackHolds(g, lagrellaFollowUpLabel); got != 1 {
		t.Fatalf("one follow-up trigger on the stack, got %d", got)
	}
	lagrellaSettle(t, g)
	if got := lagrellaReturned(t, g, "My Bear").Counters[game.CounterPlusOne]; got != 2 {
		t.Errorf("my returned creature has %d +1/+1 counters, want 2", got)
	}
	if got := lagrellaReturned(t, g, "Their Bear").Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("the opponent's returned creature has %d +1/+1 counters, want 0", got)
	}

	// The opponent's follow-up never triggered; it ends with the turn and
	// can never fire for a later entry of the same card.
	for i := 0; i < 32 && g.Turn.Seq == 1; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if n := lagrellaFollowUps(g); n != 0 {
		t.Errorf("%d follow-up triggers are still queued after the turn ended", n)
	}
}

// A creature I control but an opponent owns comes back under its owner's
// control (CR 610.3c), not mine, so it gets no counters.
func TestLagrellaStolenCreatureReturnsToItsOwnerWithoutCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stolen Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	lagrella, err := lagrellaEnters(t, g, stolen)
	if err != nil {
		t.Fatalf("picking the stolen creature: %v", err)
	}
	lagrellaSettle(t, g)
	if !g.Exile.Contains(stolen) {
		t.Fatal("the stolen creature is exiled")
	}
	destroyNow(t, g, lagrella)
	g.RunStateChecksForTest()
	lagrellaSettle(t, g)
	c := lagrellaReturned(t, g, "Stolen Bear")
	if c.Controller != opp.ID {
		t.Fatalf("the card returns under its owner's control, got %v", c.Controller)
	}
	if got := c.Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("it got %d +1/+1 counters entering under another player's control", got)
	}
}

// The targets must be controlled by different players, and Lagrella is
// not one of them ("other").
func TestLagrellaTargetsOtherCreaturesControlledByDifferentPlayers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, opp.ID, "Bear A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "Bear B", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	lagrella := castCatalogSpell(t, g, "Lagrella, the Magpie", "Legendary Creature — Human Soldier", lagrellaOracle, nil)
	passPriorityAroundTable(t, g)
	ch := lagrellaPickPrompt(g, me.ID)
	if ch == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range ch.PickTargetCards {
		if id == lagrella {
			t.Error("Lagrella offers itself as a target")
		}
	}
	err := g.ResolvePickTargets(ch.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	if err == nil {
		t.Fatal("two creatures of one player were accepted")
	}
	if err := g.ResolvePickTargets(ch.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: lagrella}}); err == nil {
		t.Fatal("Lagrella was accepted as its own target")
	}
	// Any number includes none.
	if err := g.ResolvePickTargets(ch.ID, me.ID, nil); err != nil {
		t.Fatalf("no targets is a legal answer: %v", err)
	}
	lagrellaSettle(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) || !g.Battlefield.Contains(lagrella) {
		t.Error("nothing is exiled with no targets")
	}
}

// Three creatures of three players, one of them mine, are all exiled and
// return together (CR 610.3d): only mine gets counters, so exactly one
// follow-up trigger goes on the stack.
func TestLagrellaThreeTargetsOneFollowUp(t *testing.T) {
	g := newCatalogGame(t)
	me, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	bs := pushVanillaCreature(g, b.ID, "B's Bear", 2, 2)
	cs := pushVanillaCreature(g, c.ID, "C's Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	lagrella, err := lagrellaEnters(t, g, mine, bs, cs)
	if err != nil {
		t.Fatalf("three creatures of three players: %v", err)
	}
	lagrellaSettle(t, g)
	for _, id := range []uuid.UUID{mine, bs, cs} {
		if !g.Exile.Contains(id) {
			t.Fatal("every target is exiled")
		}
	}
	destroyNow(t, g, lagrella)
	g.RunStateChecksForTest()
	if got := lagrellaStackHolds(g, lagrellaFollowUpLabel); got != 1 {
		t.Fatalf("one follow-up trigger, got %d", got)
	}
	lagrellaSettle(t, g)
	if lagrellaReturned(t, g, "My Bear").Counters[game.CounterPlusOne] != 2 ||
		lagrellaReturned(t, g, "B's Bear").Counters[game.CounterPlusOne] != 0 ||
		lagrellaReturned(t, g, "C's Bear").Counters[game.CounterPlusOne] != 0 {
		t.Error("only my creature gets the counters")
	}
}

// A Lagrella that left before its trigger resolved exiles nothing (CR
// 610.3b), so nothing returns and no counters are placed.
func TestLagrellaGoneBeforeResolutionExilesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	lagrella, err := lagrellaEnters(t, g, mine, theirs)
	if err != nil {
		t.Fatal(err)
	}
	destroyNow(t, g, lagrella)
	lagrellaSettle(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Fatal("nothing is exiled once Lagrella has left")
	}
	if counterOnBattlefieldCard(g, mine, game.CounterPlusOne) != 0 {
		t.Error("no counters without the return")
	}
}

// The return record and its follow-up are data: a table holding an exiled
// card is a restore point, and the restored game still puts the counters
// on as Lagrella leaves.
func TestLagrellaFollowUpSurvivesSnapshotAndClone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	lagrella, err := lagrellaEnters(t, g, mine)
	if err != nil {
		t.Fatal(err)
	}
	lagrellaSettle(t, g)
	if !g.Exile.Contains(mine) {
		t.Fatal("the creature is exiled")
	}
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a table holding Lagrella's card is not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	for _, next := range []*game.Game{restored, g.Clone()} {
		destroyNow(t, next, lagrella)
		next.RunStateChecksForTest()
		lagrellaSettle(t, next)
		if got := lagrellaReturned(t, next, "My Bear").Counters[game.CounterPlusOne]; got != 2 {
			t.Errorf("after a restore or clone the returned creature has %d +1/+1 counters, want 2", got)
		}
	}
}
