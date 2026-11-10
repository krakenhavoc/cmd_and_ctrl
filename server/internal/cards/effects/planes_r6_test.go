package effects

import (
	"errors"
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"testing"
)

// planes_r6_test.go — Planechase plane cards, redo 6 (ADR 0216 slice 4,
// #4870), on the real Planechase table of planes_harness_test.go.

var (
	planeIzzetSteamMaze = planeFixture{"e5f8b2bb-e9bb-4a0a-aa88-7fa0299f52a9", "Izzet Steam Maze", "Plane — Ravnica"}
)

func TestIzzetSteamMazeCopiesAnyPlayersInstantOrSorcery(t *testing.T) {
	requirePlaneRegisteredFull(t, planeIzzetSteamMaze)
	g, plane := newPlaneGame(t, planeIzzetSteamMaze)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	if got := planeTriggerControllers(g, plane.InstanceID); len(got) != 1 {
		t.Fatalf("the cast triggered %d plane items, want 1", len(got))
	}
	// Resolve the trigger only: the copy sits above the original.
	for i := 0; i < 8 && itemsFromPlane(g, plane.InstanceID) > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if n := g.Stack.Size(); n != 2 {
		t.Fatalf("stack holds %d spells after the trigger, want the original and its copy", n)
	}
	passPriorityAroundTable(t, g)

	// A creature spell is not copied.
	castCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Errorf("a creature spell triggered the plane (%d items)", n)
	}
	passPriorityAroundTable(t, g)
}

func TestIzzetSteamMazeChaosDiscountsInstantsAndSorceries(t *testing.T) {
	g, plane := newPlaneGame(t, planeIzzetSteamMaze)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the chaos ability")
	}
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if len(nsPromises(me)) != 1 {
		t.Fatalf("promises after chaos: %+v", nsPromises(me))
	}
	ritual := nsHandCard(me, "Costly Ritual", "Sorcery", "{4}{R}")
	if got := nsPrice(t, g, me, ritual).Generic; got != 1 {
		t.Errorf("a {4}{R} sorcery has %d generic after chaos, want 1", got)
	}
	bear := nsHandCard(me, "Costly Bear", "Creature — Bear", "{4}{R}")
	if got := nsPrice(t, g, me, bear).Generic; got != 4 {
		t.Errorf("a creature spell has %d generic, want 4", got)
	}
	// It outlives the plane: planeswalking away keeps the promise.
	g.WithWriteLock(func() { _ = g.PlaneswalkForEffect(me.ID) })
	if len(nsPromises(me)) != 1 {
		t.Error("the promise ended with the plane")
	}
}

var planeKilnspireDistrict = planeFixture{"fc650404-991c-4639-b311-c35df81a254a", "Kilnspire District", "Plane — Ravnica"}

func TestKilnspireDistrictChargesAndAddsManaOnEntryAndFirstMain(t *testing.T) {
	requirePlaneRegisteredFull(t, planeKilnspireDistrict)
	g, plane := newPlaneGame(t, planeKilnspireDistrict)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	me.ManaPool = game.ManaPool{}
	planeswalkToPlane(t, g, planeKilnspireDistrict.Oracle)
	passPriorityAroundTable(t, g)
	up, _ := faceUpPlane(g, planeKilnspireDistrict.Oracle)
	if up.Counters[game.CounterCharge] != 1 {
		t.Fatalf("charge counters after planeswalking to it: %d, want 1", up.Counters[game.CounterCharge])
	}
	if got := me.ManaPool.Total(); got != 1 {
		t.Errorf("mana after entry: %d, want 1", got)
	}
	_ = plane
	// Next planar controller's first main: a second counter, two mana.
	next := toNextUpkeep(t, g)
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	up, _ = faceUpPlane(g, planeKilnspireDistrict.Oracle)
	if up.Counters[game.CounterCharge] != 2 {
		t.Fatalf("charge counters after the first main phase: %d, want 2", up.Counters[game.CounterCharge])
	}
	if got := next.ManaPool.Total(); got != 2 {
		t.Errorf("mana after the first main phase: %d, want 2", got)
	}
}

func TestKilnspireDistrictChaosPaysXForDamage(t *testing.T) {
	g, plane := newPlaneGame(t, planeKilnspireDistrict)
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me.ManaPool = game.ManaPool{}
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the chaos ability")
	}
	fillPool(me, 3)
	life := opp.Life
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	answerNumber(t, g, me.ID, 3)
	pay := payUnlessFor(g, me.ID)
	if pay == nil {
		t.Fatal("no payment prompt for X")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("pay: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the target lost %d life, want 3", got)
	}
}

var planeMinamo = planeFixture{"6491f511-87e6-4504-84f5-fe5cf44768d7", "Minamo", "Plane — Kamigawa"}

func TestMinamoLetsTheCasterDrawAndChaosReturnsBlueCards(t *testing.T) {
	requirePlaneRegisteredFull(t, planeMinamo)
	g, plane := newPlaneGame(t, planeMinamo)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	// A spell's caster may draw; it resolves before the spell.
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Test Ritual", "Sorcery", "", nil)
	if n := itemsFromPlane(g, plane.InstanceID); n != 1 {
		t.Fatalf("the cast put %d plane items up, want 1", n)
	}
	for i := 0; i < 8 && itemsFromPlane(g, plane.InstanceID) > 0; i++ {
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("pass: %v", err)
		}
	}
	c := latestConfirmFor(g, me.ID)
	if c == nil {
		t.Fatal("the caster was not asked whether to draw")
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (the spell left, a card was drawn)", got, before+1)
	}
	passPriorityAroundTable(t, g)

	// Chaos: each player may return a blue card; all return together.
	rollBlank(t, g, me.ID)
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatalf("a blank triggered the chaos ability")
	}
	blue := game.Card{InstanceID: uuid.New(), Name: "Blue One", TypeLine: "Instant", Colors: []string{"U"}, Owner: me.ID, Controller: me.ID}
	me.Graveyard.PushTop(blue)
	red := pushGraveyardCardWithTypeLineID(me, "Red One", "Instant")
	theirs := game.Card{InstanceID: uuid.New(), Name: "Their Blue", TypeLine: "Instant", Colors: []string{"U"}, Owner: opp.ID, Controller: opp.ID}
	opp.Graveyard.PushTop(theirs)
	rollChaos(t, g, me.ID)
	passPriorityAroundTable(t, g)
	c = latestOptionPickFor(g, me.ID)
	if c == nil || len(c.PickOptions) != 2 {
		t.Fatalf("the active player's offer: %+v (only the blue card, plus declining)", c)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 1); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if me.Hand.Contains(blue.InstanceID) {
		t.Error("the card returned before the other players chose")
	}
	c = latestOptionPickFor(g, opp.ID)
	if c == nil {
		t.Fatal("the next player was not asked")
	}
	if err := g.ResolveOptionPick(c.ID, opp.ID, 0); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(blue.InstanceID) {
		t.Error("the chosen blue card did not return")
	}
	if me.Hand.Contains(red) || opp.Hand.Contains(theirs.InstanceID) {
		t.Error("a card that was not chosen returned")
	}
}
