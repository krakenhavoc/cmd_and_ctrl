package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ugins_nexus_test.go — "If a player would begin an extra turn, that
// player skips that turn instead" and "if Ugin's Nexus would be put into
// a graveyard from the battlefield, instead exile it and take an extra
// turn after this one" (ADR 0188 Decision 2, #3560, CR 614.10). Each test
// drives the real rotation seam and asserts who is active afterwards.

const (
	uginsNexusOracle      = "fe2afa18-54d0-4595-af27-256398793a42"
	hourglassPendantOracl = "01591d42-4ef8-471d-beda-d31186bba6a6"
)

func seedNexus(g *game.Game, controller uuid.UUID) uuid.UUID {
	return seedEnchantment(g, uginsNexusOracle, "Ugin's Nexus", "Legendary Artifact", controller)
}

func queueOwnExtraTurn(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", temporalManipulationOracle, nil)
	passPriorityAroundTable(t, g)
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Fatalf("the spell should queue its turn: %d queued", got)
	}
}

func TestUginsNexusAndPendantAreRegisteredFull(t *testing.T) {
	for oracle, want := range map[string]string{
		uginsNexusOracle:      "Ugin's Nexus",
		hourglassPendantOracl: "Gerrard's Hourglass Pendant",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want)
			continue
		}
		if spec.Name != want || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q completeness %v", want, spec.Name, spec.Completeness)
		}
	}
}

// "A player": the Nexus controller's OWN extra turn is skipped too, unlike
// Trouble in Pairs.
func TestUginsNexusSkipsItsControllersOwnExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	seedNexus(g, g.Seats[me].ID)
	begun := g.Seats[me].TurnsBegun

	queueOwnExtraTurn(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the controller's skipped extra turn")
	if got := skippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
	if got := g.Seats[me].TurnsBegun; got != begun {
		t.Errorf("a skipped turn counted as begun: TurnsBegun %d, want %d", got, begun)
	}
}

// And an opponent's, whoever's Nexus it is.
func TestUginsNexusSkipsAnOpponentsExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	opp := (me + 2) % len(g.Seats)
	seedNexus(g, g.Seats[me].ID)

	castCatalogSpell(t, g, "Time Warp", "Sorcery", timeWarpOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[opp].ID}})
	passPriorityAroundTable(t, g)
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the opponent's skipped extra turn")
	if got := skippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
}

// The ruling: the skip is judged when the turn would begin. A turn queued
// while the Nexus is on the battlefield is taken if the Nexus is gone by
// then. The Nexus is bounced, not destroyed, so no turn of its own.
func TestExtraTurnQueuedUnderTheNexusIsTakenIfTheNexusLeft(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	nexus := seedNexus(g, g.Seats[me].ID)

	queueOwnExtraTurn(t, g)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneHand, Owner: g.Seats[me].ID}, nexus); err != nil {
		t.Fatalf("bounce the Nexus: %v", err)
	}
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "extra turn queued under a Nexus that left")
	if got := skippedExtraTurnEvents(g); got != 0 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 0", got)
	}
}

// The other direction: a turn queued before the Nexus arrived is skipped
// when it would begin.
func TestExtraTurnQueuedBeforeTheNexusArrivedIsSkipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	queueOwnExtraTurn(t, g)
	seedNexus(g, g.Seats[me].ID)
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "extra turn queued before the Nexus arrived")
}

// Destroying the Nexus exiles it (it never reaches a graveyard) and
// queues one extra turn for its controller, which is taken because the
// Nexus is no longer there to skip it. The exile is logged before the
// extra turn.
func TestDestroyingTheNexusExilesItAndTakesAnExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	nexus := seedNexus(g, g.Seats[me].ID)

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

	endTurn(t, g)
	assertTurnOf(t, g, me, true, "the extra turn the Nexus gave")
	if got := skippedExtraTurnEvents(g); got != 0 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 0", got)
	}
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "rotation resumes after the extra turn")
}

// A second Nexus on the battlefield still skips the turn the first
// one's death gave: it is skipped only because a Nexus is still there when
// it would begin (the ruling), and not at all when none is.
func TestAnotherNexusStillSkipsTheTurnTheFirstOneGave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	first := seedNexus(g, g.Seats[me].ID)
	second := seedNexus(g, g.Seats[me].ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(first) })
	if !g.Exile.Contains(first) || !g.Battlefield.Contains(second) {
		t.Fatalf("the first Nexus should be exiled and the second remain (exile %v, bf %v)",
			g.Exile.Contains(first), g.Battlefield.Contains(second))
	}
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Fatalf("%d extra turns queued, want 1", got)
	}
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "the remaining Nexus skips the turn")
	if got := skippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
}

// And if the remaining Nexus leaves before the turn begins, the turn is
// taken.
func TestTheTurnTheNexusGaveIsTakenOnceTheOtherNexusLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	first := seedNexus(g, g.Seats[me].ID)
	second := seedNexus(g, g.Seats[me].ID)

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(first) })
	// The second one is destroyed too, which gives its own turn.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(second) })
	if got := len(g.ExtraTurnsQueued()); got != 2 {
		t.Fatalf("%d extra turns queued, want 2", got)
	}
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "first extra turn")
	endTurn(t, g)
	assertTurnOf(t, g, me, true, "second extra turn")
	endTurn(t, g)
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "rotation resumes")
}

// Two Nexuses destroyed together: each replacement applies to its own
// move, so both are exiled and each queues a turn.
func TestTwoNexusesDestroyedTogetherAreBothExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	a := seedNexus(g, g.Seats[me].ID)
	b := seedNexus(g, g.Seats[me].ID)

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
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	// Rest in Peace belongs to ANOTHER seat: the prompt must still go
	// to the Nexus' controller.
	seedEnchantment(g, restInPeaceOracle, "Rest in Peace", "Enchantment", g.Seats[(me+1)%len(g.Seats)].ID)
	nexus := seedNexus(g, g.Seats[me].ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(nexus) })
	return g, nexus, me
}

func answerNexusOrder(t *testing.T, g *game.Game, firstLabelContains string, wantChooser uuid.UUID) {
	t.Helper()
	var prompt *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceReplacementOrder {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatalf("no CR 616 ordering prompt (have %d choices)", len(g.PendingChoices))
	}
	if prompt.Chooser != wantChooser {
		t.Fatalf("the order is asked of %v, want the Nexus' controller %v", prompt.Chooser, wantChooser)
	}
	if len(prompt.ReplacementEffectIDs) != 2 {
		t.Fatalf("%d replacements in the prompt, want 2", len(prompt.ReplacementEffectIDs))
	}
	labels := map[game.ReplacementEffectID]string{}
	g.WithWriteLock(func() {
		for _, id := range prompt.ReplacementEffectIDs {
			labels[id], _ = g.ReplacementOptionMetaForEffect(id)
		}
	})
	var ids []game.ReplacementEffectID
	for id, l := range labels {
		if containsFold(l, firstLabelContains) {
			ids = append([]game.ReplacementEffectID{id}, ids...)
		} else {
			ids = append(ids, id)
		}
	}
	if len(ids) != 2 || !containsFold(labels[ids[0]], firstLabelContains) {
		t.Fatalf("no option mentioning %q in %v", firstLabelContains, labels)
	}
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, ids); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
}

func TestNexusWithRestInPeaceNexusFirstGivesTheTurn(t *testing.T) {
	g, nexus, me := nexusRestInPeaceBoard(t)
	answerNexusOrder(t, g, "Ugin's Nexus", g.Seats[me].ID)
	if !g.Exile.Contains(nexus) {
		t.Fatal("the Nexus is not in exile")
	}
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Fatalf("%d extra turns queued, want 1 (Nexus first)", got)
	}
}

func TestNexusWithRestInPeaceRestInPeaceFirstGivesNoTurn(t *testing.T) {
	g, nexus, me := nexusRestInPeaceBoard(t)
	answerNexusOrder(t, g, "Rest in Peace", g.Seats[me].ID)
	if !g.Exile.Contains(nexus) {
		t.Fatal("the Nexus is not in exile")
	}
	if got := len(g.ExtraTurnsQueued()); got != 0 {
		t.Fatalf("%d extra turns queued, want 0 (Rest in Peace first)", got)
	}
}
