package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestUginsNexusAndPendantAreRegisteredFull(t *testing.T) {
	for oracle, want := range map[string]string{
		effectstest.UginsNexusOracle:      "Ugin's Nexus",
		effectstest.HourglassPendantOracl: "Gerrard's Hourglass Pendant",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want)
			continue
		}
		if spec.Name != want || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q completeness %v", want, spec.Name, spec.Completeness)
		}
	}
}

// "A player": the Nexus controller's OWN extra turn is skipped too, unlike
// Trouble in Pairs.
func TestUginsNexusSkipsItsControllersOwnExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	effectstest.SeedNexus(g, g.Seats[me].ID)
	begun := g.Seats[me].TurnsBegun

	effectstest.QueueOwnExtraTurn(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the controller's skipped extra turn")
	if got := effectstest.SkippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
	if got := g.Seats[me].TurnsBegun; got != begun {
		t.Errorf("a skipped turn counted as begun: TurnsBegun %d, want %d", got, begun)
	}
}

// And an opponent's, whoever's Nexus it is.
func TestUginsNexusSkipsAnOpponentsExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 2) % len(g.Seats)
	effectstest.SeedNexus(g, g.Seats[me].ID)

	effectstest.CastCatalogSpell(t, g, "Time Warp", "Sorcery", effectstest.TimeWarpOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the opponent's skipped extra turn")
	if got := effectstest.SkippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
}

// The ruling: the skip is judged when the turn would begin. A turn queued
// while the Nexus is on the battlefield is taken if the Nexus is gone by
// then. The Nexus is bounced, not destroyed, so no turn of its own.
func TestExtraTurnQueuedUnderTheNexusIsTakenIfTheNexusLeft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	nexus := effectstest.SeedNexus(g, g.Seats[me].ID)

	effectstest.QueueOwnExtraTurn(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneHand, Owner: g.Seats[me].ID}, nexus); err != nil {
		t.Fatalf("bounce the Nexus: %v", err)
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "extra turn queued under a Nexus that left")
	if got := effectstest.SkippedExtraTurnEvents(g); got != 0 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 0", got)
	}
}

// The other direction: a turn queued before the Nexus arrived is skipped
// when it would begin.
func TestExtraTurnQueuedBeforeTheNexusArrivedIsSkipped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	effectstest.QueueOwnExtraTurn(t, g)
	effectstest.SeedNexus(g, g.Seats[me].ID)
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "extra turn queued before the Nexus arrived")
}

// Destroying the Nexus exiles it (it never reaches a graveyard) and
// queues one extra turn for its controller, which is taken because the
// Nexus is no longer there to skip it. The exile is logged before the
// extra turn.
func TestDestroyingTheNexusExilesItAndTakesAnExtraTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	nexus := effectstest.SeedNexus(g, g.Seats[me].ID)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(nexus); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(nexus) {
		t.Fatal("the Nexus is not in exile")
	}
	if g.Seats[me].Graveyard.Contains(nexus) {
		t.Fatal("the Nexus reached the graveyard")
	}
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Fatalf("%d extra turns queued, want 1", got)
	}

	// Event order: the move to exile, then the extra turn.
	moveAt, turnAt := -1, -1
	for i, ev := range g.Events {
		switch {
		case ev.Kind == game.EventZoneMove && ev.CardID == nexus && ev.NewZone == game.ZoneExile:
			moveAt = i
		case ev.Kind == game.EventExtraTurnAdded:
			turnAt = i
		}
	}
	if moveAt < 0 || turnAt < 0 || turnAt < moveAt {
		t.Errorf("exile at event %d, extra turn at %d: the turn must follow the exile", moveAt, turnAt)
	}
	// And the same order on the wire log.
	exileLine, turnLine := -1, -1
	for i, e := range protocol.ViewOfGame(g).Log {
		if e.Kind == protocol.LogExtraTurn {
			turnLine = i
		}
		if e.CardID == nexus.String() && e.Kind != protocol.LogExtraTurn {
			exileLine = i
		}
	}
	if turnLine < 0 {
		t.Fatal("no extra_turn log line")
	}
	if exileLine < 0 || exileLine > turnLine {
		t.Errorf("the exile line (%d) must come before the extra_turn line (%d)", exileLine, turnLine)
	}

	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "the extra turn the Nexus gave")
	if got := effectstest.SkippedExtraTurnEvents(g); got != 0 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 0", got)
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "rotation resumes after the extra turn")
}

// A second Nexus on the battlefield still skips the turn the first
// one's death gave: it is skipped only because a Nexus is still there when
// it would begin (the ruling), and not at all when none is.
func TestAnotherNexusStillSkipsTheTurnTheFirstOneGave(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	first := effectstest.SeedNexus(g, g.Seats[me].ID)
	second := effectstest.SeedNexus(g, g.Seats[me].ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(first) })
	if !g.Exile.Contains(first) || !g.Battlefield.Contains(second) {
		t.Fatalf("the first Nexus should be exiled and the second remain (exile %v, bf %v)",
			g.Exile.Contains(first), g.Battlefield.Contains(second))
	}
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Fatalf("%d extra turns queued, want 1", got)
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "the remaining Nexus skips the turn")
	if got := effectstest.SkippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
}

// And if the remaining Nexus leaves before the turn begins, the turn is
// taken.
func TestTheTurnTheNexusGaveIsTakenOnceTheOtherNexusLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	first := effectstest.SeedNexus(g, g.Seats[me].ID)
	second := effectstest.SeedNexus(g, g.Seats[me].ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(first) })
	// The second one is destroyed too, which gives its own turn.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(second) })
	if got := len(g.ExtraTurnsQueued()); got != 2 {
		t.Fatalf("%d extra turns queued, want 2", got)
	}
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "first extra turn")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, me, true, "second extra turn")
	effectstest.EndTurn(t, g)
	effectstest.AssertTurnOf(t, g, (me+1)%len(g.Seats), false, "rotation resumes")
}

// Two Nexuses destroyed together: each replacement applies to its own
// move, so both are exiled and each queues a turn.
func TestTwoNexusesDestroyedTogetherAreBothExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	a := effectstest.SeedNexus(g, g.Seats[me].ID)
	b := effectstest.SeedNexus(g, g.Seats[me].ID)

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentsKeyedForEffect([]uuid.UUID{a, b}, game.KeyedTail{}); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	for _, id := range []uuid.UUID{a, b} {
		if !g.Exile.Contains(id) {
			t.Errorf("a Nexus is not in exile")
		}
	}
	if got := len(g.ExtraTurnsQueued()); got != 2 {
		t.Fatalf("%d extra turns queued, want 2", got)
	}
}

// Rest in Peace is a second graveyard replacement on the same move. The
// Nexus' controller chooses the order (CR 616.1). Rest in Peace first
// exiles the Nexus without the extra turn; the Nexus first gives the
// turn.
func nexusRestInPeaceBoard(t *testing.T) (*game.Game, uuid.UUID, int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Turn.ActiveSeat
	// Rest in Peace belongs to ANOTHER seat: the prompt must still go
	// to the Nexus' controller.
	effectstest.SeedEnchantment(g, effectstest.RestInPeaceOracle, "Rest in Peace", "Enchantment", g.Seats[(me+1)%len(g.Seats)].ID)
	nexus := effectstest.SeedNexus(g, g.Seats[me].ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(nexus) })
	return g, nexus, me
}

func TestNexusWithRestInPeaceNexusFirstGivesTheTurn(t *testing.T) {
	g, nexus, me := nexusRestInPeaceBoard(t)
	effectstest.AnswerNexusOrder(t, g, "Ugin's Nexus", g.Seats[me].ID)
	if !g.Exile.Contains(nexus) {
		t.Fatal("the Nexus is not in exile")
	}
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Fatalf("%d extra turns queued, want 1 (Nexus first)", got)
	}
}

func TestNexusWithRestInPeaceRestInPeaceFirstGivesNoTurn(t *testing.T) {
	g, nexus, me := nexusRestInPeaceBoard(t)
	effectstest.AnswerNexusOrder(t, g, "Rest in Peace", g.Seats[me].ID)
	if !g.Exile.Contains(nexus) {
		t.Fatal("the Nexus is not in exile")
	}
	if got := len(g.ExtraTurnsQueued()); got != 0 {
		t.Fatalf("%d extra turns queued, want 0 (Rest in Peace first)", got)
	}
}
