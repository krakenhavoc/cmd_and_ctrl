package protocol

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// game_end_view_test.go — ADR 0057 Decision 7 (#749), test plan item
// 17: the outcome on the wire, the log's one line per departure, the
// game_over and win_prevented lines, and the cant_lose seat fields.

const testGateOracle = "test-749-view-angel"

func withViewGates(t *testing.T) {
	t.Helper()
	prev := game.CatalogGameEndGates
	game.CatalogGameEndGates = func(key string) []game.GameEndGate {
		if key == testGateOracle {
			return []game.GameEndGate{
				{Scope: game.GateYou, CantLose: true},
				{Scope: game.GateOpponents, CantWin: true},
			}
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogGameEndGates = prev })
}

func seedViewAngel(g *game.Game, owner *game.Player) uuid.UUID {
	c := game.NewCard("Platinum Angel", owner.ID)
	c.OracleID = testGateOracle
	c.TypeLine = "Artifact Creature — Angel"
	c.Power, c.Toughness = 4, 4
	g.WithWriteLock(func() { g.Battlefield.PushTop(c) })
	return c.InstanceID
}

func logOfKind(v GameView, kind LogKind) []LogEvent {
	var out []LogEvent
	for _, e := range v.Log {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}

// TestConcedeLogsOneLineAndAnOutcome: one concession, one eliminated
// line (the engine's concede event is silent), and a two-seat game's
// outcome names the survivor with a game_over line.
func TestConcedeLogsOneLineAndAnOutcome(t *testing.T) {
	g := newTwoSeatGame(t)
	a, b := g.Seats[0], g.Seats[1]
	if err := g.Concede(a.ID); err != nil {
		t.Fatal(err)
	}
	v := ViewOfGame(g)
	elim := logOfKind(v, LogEliminated)
	if len(elim) != 1 {
		t.Fatalf("%d eliminated lines for one concession, want 1: %+v", len(elim), elim)
	}
	if elim[0].Cause != "concede" || elim[0].Amount != 1 || elim[0].Text != "A conceded" {
		t.Errorf("eliminated line %+v", elim[0])
	}
	if v.Outcome == nil || v.Outcome.Kind != "win" || v.Outcome.Winner != b.ID.String() ||
		v.Outcome.WinnerSeat == nil || *v.Outcome.WinnerSeat != b.Seat || v.Outcome.Cause != "last_standing" {
		t.Fatalf("outcome %+v", v.Outcome)
	}
	over := logOfKind(v, LogGameOver)
	if len(over) != 1 || over[0].Text != "B won the game: every opponent has left" {
		t.Errorf("game_over lines %+v", over)
	}
	// Filtered views carry it too, for every seat.
	if fv := FilterViewFor(v, a.ID.String()); fv.Outcome == nil || fv.Outcome.Winner != b.ID.String() {
		t.Errorf("filtered outcome %+v", fv.Outcome)
	}
}

// TestEffectWinOutcomeNamesItsSource: an effect win's outcome and log
// line name the winning source, and the other seat is still seated.
func TestEffectWinOutcomeNamesItsSource(t *testing.T) {
	g := newTwoSeatGame(t)
	a := g.Seats[0]
	src := game.NewCard("Felidar Sovereign", a.ID)
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(src)
		if _, err := g.WinTheGameForEffect(a.ID, src.InstanceID); err != nil {
			t.Fatal(err)
		}
	})
	v := ViewOfGame(g)
	if v.Outcome == nil || v.Outcome.Cause != "effect" || v.Outcome.SourceName != "Felidar Sovereign" ||
		v.Outcome.Source != src.InstanceID.String() {
		t.Fatalf("outcome %+v", v.Outcome)
	}
	if v.Seats[1].Eliminated {
		t.Errorf("the other seat is eliminated by an effect win")
	}
	over := logOfKind(v, LogGameOver)
	if len(over) != 1 || over[0].Text != "A won the game (Felidar Sovereign)" {
		t.Errorf("game_over lines %+v", over)
	}
}

// TestDrawOutcomeHasNoWinner: CR 104.4a on the wire.
func TestDrawOutcomeHasNoWinner(t *testing.T) {
	g := newTwoSeatGame(t)
	g.WithWriteLock(func() {
		g.Seats[0].Life = 0
		g.Seats[1].Life = 0
	})
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	v := ViewOfGame(g)
	if v.Outcome == nil || v.Outcome.Kind != "draw" || v.Outcome.Winner != "" || v.Outcome.WinnerSeat != nil {
		t.Fatalf("outcome %+v", v.Outcome)
	}
	over := logOfKind(v, LogGameOver)
	if len(over) != 1 || over[0].Text != "The game is a draw" || over[0].Seat != NoSeat {
		t.Errorf("game_over lines %+v", over)
	}
	elim := logOfKind(v, LogEliminated)
	if len(elim) != 2 || !strings.HasSuffix(elim[0].Text, "lost the game (0 or less life)") {
		t.Errorf("eliminated lines %+v", elim)
	}
}

// TestEndHasNoOutcomeOnTheWire: an admin ending a table is not a
// result.
func TestEndHasNoOutcomeOnTheWire(t *testing.T) {
	g := newTwoSeatGame(t)
	g.End()
	if v := ViewOfGame(g); v.Outcome != nil || v.State != "ended" {
		t.Fatalf("state %s outcome %+v", v.State, v.Outcome)
	}
}

// TestGatedSeatCarriesCantLoseAndTheOpponentCantWin, and a prevented
// win writes its line.
func TestGatedSeatCarriesCantLoseAndTheOpponentCantWin(t *testing.T) {
	withViewGates(t)
	g := newTwoSeatGame(t)
	a, b := g.Seats[0], g.Seats[1]
	angel := seedViewAngel(g, a)
	lab := game.NewCard("Laboratory Maniac", b.ID)
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(lab)
		if won, _ := g.WinTheGameForEffect(b.ID, lab.InstanceID); won {
			t.Fatalf("won past the Angel")
		}
	})
	v := ViewOfGame(g)
	sa, sb := v.Seats[0], v.Seats[1]
	if strings.Join(sa.CantLose, ",") != "life,empty_draw,poison,commander_damage,effect" || sa.CantWin {
		t.Errorf("seat A cant_lose %v cant_win %v", sa.CantLose, sa.CantWin)
	}
	if len(sa.EndGates) != 1 || sa.EndGates[0].SourceName != "Platinum Angel" || sa.EndGates[0].Source != angel.String() {
		t.Errorf("seat A end_gates %+v", sa.EndGates)
	}
	if len(sb.CantLose) != 0 || !sb.CantWin {
		t.Errorf("seat B cant_lose %v cant_win %v", sb.CantLose, sb.CantWin)
	}
	prevented := logOfKind(v, LogWinPrevented)
	if len(prevented) != 1 || prevented[0].Text != "B would have won the game (Laboratory Maniac), but can't (Platinum Angel)" {
		t.Errorf("win_prevented lines %+v", prevented)
	}
}

// TestEffectLossLineNamesTheSource: "Alice lost the game (Pact of
// Negation)".
func TestEffectLossLineNamesTheSource(t *testing.T) {
	// Three seats, so the loss does not end the game.
	g := newSeatsGame(t, 3)
	loser := g.Seats[1]
	pact := game.NewCard("Pact of Negation", loser.ID)
	holder := g.Seats[0]
	pact.Owner = holder.ID
	g.WithWriteLock(func() {
		holder.Graveyard.PushTop(pact)
		if _, err := g.LoseTheGameForEffect(loser.ID, pact.InstanceID); err != nil {
			t.Fatal(err)
		}
	})
	v := ViewOfGame(g)
	elim := logOfKind(v, LogEliminated)
	if len(elim) != 1 || elim[0].Cause != "effect" || elim[0].Text != "B lost the game (Pact of Negation)" {
		t.Errorf("eliminated lines %+v", elim)
	}
	if v.Outcome != nil {
		t.Errorf("a three-seat game ended on one loss")
	}
}

// newSeatsGame is newTwoSeatGame for any table size.
func newSeatsGame(t *testing.T, n int) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < n; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer %d: %v", i, err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}
