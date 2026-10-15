package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_skip_order_test.go — asking which skip takes an extra turn (ADR
// 0188 Decision 12, CR 616.1, #3561). A player who owes "skip your next
// turn" (Magosi, the Waterveil) and whose extra turn an extra-turn skip
// (Trouble in Pairs, Ugin's Nexus) would also skip is asked which effect
// takes it. Each test drives the real cast and the real rotation seam
// and asserts who is active once the question is answered.

// openTurnSkipOrder returns the open which-skip question, failing the
// test unless it is the only prompt.
func openTurnSkipOrder(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != game.PendingChoiceTurnSkipOrder {
		t.Fatalf("want one turn_skip_order prompt, have %d prompts", len(g.PendingChoices))
	}
	return g.PendingChoices[0]
}

// owesMagosiSkip writes Magosi's "skip your next turn" for `seat`, with a
// Magosi on the battlefield as its source.
func owesMagosiSkip(t *testing.T, g *game.Game, seat int) uuid.UUID {
	t.Helper()
	magosi := magosiOnBattlefield(g, g.Seats[seat].ID)
	g.WithWriteLock(func() {
		if !g.SkipNextTurnsForEffect(g.Seats[seat].ID, magosi, 1, "Magosi — skip your next turn") {
			t.Fatal("no skip recorded")
		}
	})
	return magosi
}

// endTurnIntoTheQuestion ends the turn and asserts the turn change
// paused: Seq unchanged, the old turn's seat still active, the flag set.
func endTurnIntoTheQuestion(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	seat, seq := g.Turn.ActiveSeat, g.Turn.Seq
	endTurn(t, g)
	if g.Turn.Seq != seq || g.Turn.ActiveSeat != seat || !g.TurnChangePending {
		t.Fatalf("the turn change did not pause: seq %d (was %d), seat %d (was %d), pending %v",
			g.Turn.Seq, seq, g.Turn.ActiveSeat, seat, g.TurnChangePending)
	}
	return openTurnSkipOrder(t, g)
}

// troubleAndMagosiDebt is the acceptance board: the active player owes
// Magosi's skip and queues an extra turn (Temporal Manipulation), and the
// next seat's Trouble in Pairs would skip that turn too. The turn is
// ended, so the question is open.
func troubleAndMagosiDebt(t *testing.T) (g *game.Game, caster, other int, magosi uuid.UUID, c *game.PendingChoice) {
	t.Helper()
	g = newCatalogGame(t)
	caster = g.Turn.ActiveSeat
	other = (caster + 1) % len(g.Seats)
	seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)
	magosi = owesMagosiSkip(t, g, caster)
	castCatalogSpell(t, g, "Temporal Manipulation", "Sorcery", temporalManipulationOracle, nil)
	passPriorityAroundTable(t, g)
	return g, caster, other, magosi, endTurnIntoTheQuestion(t, g)
}

func TestMagosiAndTroubleInPairsAskWhichSkipTakesTheExtraTurn(t *testing.T) {
	g, caster, _, magosi, c := troubleAndMagosiDebt(t)
	if c.Chooser != g.Seats[caster].ID || c.Source != magosi {
		t.Errorf("chooser %v source %v, want the caster and Magosi", c.Chooser, c.Source)
	}
	if c.AcceptLabel != "Use my skip (Magosi, the Waterveil)" || c.DeclineLabel != "Let Trouble in Pairs skip it" {
		t.Errorf("labels %q / %q", c.AcceptLabel, c.DeclineLabel)
	}
	if got := len(g.ExtraTurnsQueued()); got != 1 {
		t.Errorf("%d extra turns queued while the question is open, want the one asked about", got)
	}
}

// Answer 0: Magosi's skip takes the turn. The charge is spent, Trouble
// in Pairs does nothing, and the next seat's turn begins.
func TestUsingTheMagosiSkipSpendsItOnTheExtraTurn(t *testing.T) {
	g, caster, other, _, c := troubleAndMagosiDebt(t)
	if err := g.ResolveTurnSkipOrder(c.ID, c.Chooser, true); err != nil {
		t.Fatalf("ResolveTurnSkipOrder: %v", err)
	}
	assertTurnOf(t, g, other, false, "after Magosi skips the extra turn")
	if g.TurnChangePending {
		t.Error("the turn change is still pending")
	}
	if got := g.TurnSkipsFor(g.Seats[caster].ID); got != 0 {
		t.Errorf("the charge was not spent: %d owed", got)
	}
	if got := skippedExtraTurnEvents(g); got != 0 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 0: Magosi took the turn", got)
	}
}

// Answer 1: Trouble in Pairs takes the turn, and Magosi's skip is still
// owed for the caster's next turn.
func TestLettingTroubleInPairsSkipItKeepsTheMagosiDebt(t *testing.T) {
	g, caster, other, _, c := troubleAndMagosiDebt(t)
	if err := g.ResolveTurnSkipOrder(c.ID, c.Chooser, false); err != nil {
		t.Fatalf("ResolveTurnSkipOrder: %v", err)
	}
	assertTurnOf(t, g, other, false, "after Trouble in Pairs skips the extra turn")
	if got := g.TurnSkipsFor(g.Seats[caster].ID); got != 1 {
		t.Errorf("the caster owes %d skips, want 1", got)
	}
	if got := skippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
}

// Only the player's own skip applies (no Trouble in Pairs): no question,
// the charge is spent on the extra turn (the Magosi ruling).
func TestMagosiAloneSkipsTheExtraTurnWithoutAsking(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Turn.ActiveSeat
	owesMagosiSkip(t, g, caster)
	queueOwnExtraTurn(t, g)
	endTurn(t, g)
	if len(g.PendingChoices) != 0 || g.TurnChangePending {
		t.Fatalf("asked with only one skip applying: %d prompts, pending %v", len(g.PendingChoices), g.TurnChangePending)
	}
	assertTurnOf(t, g, (caster+1)%len(g.Seats), false, "after Magosi skips the extra turn")
	if got := g.TurnSkipsFor(g.Seats[caster].ID); got != 0 {
		t.Errorf("the charge was not spent: %d owed", got)
	}
}

// Only the extra-turn skip applies (no debt): no question, Trouble in
// Pairs skips the turn as before.
func TestTroubleInPairsAloneSkipsTheExtraTurnWithoutAsking(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Turn.ActiveSeat
	other := (caster + 1) % len(g.Seats)
	seedEnchantment(g, troubleInPairsOracle, "Trouble in Pairs", "Enchantment", g.Seats[other].ID)
	queueOwnExtraTurn(t, g)
	endTurn(t, g)
	if len(g.PendingChoices) != 0 || g.TurnChangePending {
		t.Fatalf("asked with only one skip applying: %d prompts, pending %v", len(g.PendingChoices), g.TurnChangePending)
	}
	assertTurnOf(t, g, other, false, "after Trouble in Pairs skips the extra turn")
	if got := skippedExtraTurnEvents(g); got != 1 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 1", got)
	}
}

// Ugin's Nexus skips its own controller's extra turns, so a Nexus
// controller who owes Magosi's skip is asked too.
func TestMagosiAndUginsNexusAskWhichSkipTakesTheExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	seedNexus(g, g.Seats[me].ID)
	owesMagosiSkip(t, g, me)
	queueOwnExtraTurn(t, g)
	c := endTurnIntoTheQuestion(t, g)
	if c.DeclineLabel != "Let Ugin's Nexus skip it" {
		t.Errorf("decline label %q", c.DeclineLabel)
	}
	if err := g.ResolveTurnSkipOrder(c.ID, c.Chooser, false); err != nil {
		t.Fatalf("ResolveTurnSkipOrder: %v", err)
	}
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the Nexus skips the extra turn")
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 1 {
		t.Errorf("owes %d skips, want 1", got)
	}
}

// The Nexus leaves the battlefield while the question is open, and
// answer 1 is chosen: the Nexus no longer skips anything, so the Magosi
// charge is spent instead. The turn is still skipped.
func TestTheNexusLeavingWhileAskedSpendsTheMagosiChargeInstead(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Turn.ActiveSeat
	nexus := seedNexus(g, g.Seats[me].ID)
	owesMagosiSkip(t, g, me)
	queueOwnExtraTurn(t, g)
	c := endTurnIntoTheQuestion(t, g)

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(nexus); err != nil {
			t.Fatalf("exile the Nexus: %v", err)
		}
	})
	if err := g.ResolveTurnSkipOrder(c.ID, c.Chooser, false); err != nil {
		t.Fatalf("ResolveTurnSkipOrder: %v", err)
	}
	assertTurnOf(t, g, (me+1)%len(g.Seats), false, "after the extra turn is skipped by Magosi instead")
	if got := g.TurnSkipsFor(g.Seats[me].ID); got != 0 {
		t.Errorf("owes %d skips, want 0: the charge took the turn", got)
	}
	if got := skippedExtraTurnEvents(g); got != 0 {
		t.Errorf("EventExtraTurnSkipped emitted %d times, want 0: the Nexus had left", got)
	}
}
