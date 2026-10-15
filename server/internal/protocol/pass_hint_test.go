package protocol

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pass_hint_test.go — #3826: TurnView.pass_hint says, per viewer, when the
// viewer's own pass ends their turn and when it is the last one before
// their turn begins.

// passHintTable is an n-seat game past the mulligans, in seat 0's turn.
func passHintTable(t *testing.T, n int) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := range n {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// hintAt parks the cursor on a step with the given seat holding priority,
// and returns the hint seat `viewer` is shown.
func hintAt(t *testing.T, g *game.Game, step game.Step, holder, viewer int) string {
	t.Helper()
	g.WithWriteLock(func() {
		g.Turn.Step = step
		g.Turn.PriorityHolder = holder
	})
	v := FilterViewFor(ViewOfGame(g), g.Seats[viewer].ID.String())
	return v.Turn.PassHint
}

func TestPassHintOnYourOwnEndStepSaysThePassEndsTheTurn(t *testing.T) {
	g := passHintTable(t, 2)
	me := g.Turn.ActiveSeat
	if got := hintAt(t, g, game.StepEnd, me, me); got != PassHintEndsTurn {
		t.Errorf("own end step: pass_hint = %q, want %q", got, PassHintEndsTurn)
	}
	// Every earlier step is a routine pass: more of the turn is left.
	for _, step := range []game.Step{
		game.StepUpkeep, game.StepDraw, game.StepPrecombatMain, game.StepBeginCombat,
		game.StepDeclareAttackers, game.StepEndCombat, game.StepPostcombatMain,
	} {
		if got := hintAt(t, g, step, me, me); got != "" {
			t.Errorf("own %s: pass_hint = %q, want none", step, got)
		}
	}
}

func TestPassHintNeedsPriorityAndAnEmptyStack(t *testing.T) {
	g := passHintTable(t, 2)
	me := g.Turn.ActiveSeat
	other := 1 - me
	// The opponent holds priority in my end step: nothing to say to me.
	if got := hintAt(t, g, game.StepEnd, other, me); got != "" {
		t.Errorf("without priority: pass_hint = %q, want none", got)
	}
	// Cleanup grants none.
	if got := hintAt(t, g, game.StepCleanup, me, me); got != "" {
		t.Errorf("cleanup: pass_hint = %q, want none", got)
	}
	// Something on the stack: a pass resolves it and asks again.
	g.WithWriteLock(func() {
		g.Turn.Step = game.StepEnd
		g.Turn.PriorityHolder = me
	})
	v := ViewOfGame(g)
	v.Stack.Cards = append(v.Stack.Cards, CardView{Name: "Shock"})
	if got := passHintFor(v, g.Seats[me].ID.String()); got != "" {
		t.Errorf("stack not empty: pass_hint = %q, want none", got)
	}
	v.Stack.Cards = nil
	if got := passHintFor(v, g.Seats[me].ID.String()); got != PassHintEndsTurn {
		t.Errorf("stack empty: pass_hint = %q, want %q", got, PassHintEndsTurn)
	}
}

func TestPassHintInTheOpponentsEndStepBeforeYourTurn(t *testing.T) {
	g := passHintTable(t, 2)
	active := g.Turn.ActiveSeat
	me := 1 - active
	if got := hintAt(t, g, game.StepEnd, me, me); got != PassHintYourTurnNext {
		t.Errorf("opponent's end step: pass_hint = %q, want %q", got, PassHintYourTurnNext)
	}
	// Earlier in their turn you still get priority, but not for the last time.
	if got := hintAt(t, g, game.StepBeginCombat, me, me); got != "" {
		t.Errorf("opponent's beginning of combat: pass_hint = %q, want none", got)
	}
}

func TestPassHintAtAFourPlayerTableIsOnlyForTheSeatThatIsNext(t *testing.T) {
	g := passHintTable(t, 4)
	active := g.Turn.ActiveSeat
	next := (active + 1) % 4
	for viewer := range 4 {
		got := hintAt(t, g, game.StepEnd, viewer, viewer)
		want := ""
		switch viewer {
		case active:
			want = PassHintEndsTurn
		case next:
			want = PassHintYourTurnNext
		}
		if got != want {
			t.Errorf("seat %d (active %d, next %d): pass_hint = %q, want %q", viewer, active, next, got, want)
		}
	}
}

func TestPassHintFollowsASkippedTurn(t *testing.T) {
	g := passHintTable(t, 4)
	active := g.Turn.ActiveSeat
	skipper := (active + 1) % 4
	after := (active + 2) % 4
	g.WithWriteLock(func() { g.SkipNextTurnsForEffect(g.Seats[skipper].ID, uuid.Nil, 1, "test skip") })

	if got := hintAt(t, g, game.StepEnd, skipper, skipper); got != "" {
		t.Errorf("the seat that skips: pass_hint = %q, want none", got)
	}
	if got := hintAt(t, g, game.StepEnd, after, after); got != PassHintYourTurnNext {
		t.Errorf("the seat after the skipped one: pass_hint = %q, want %q", got, PassHintYourTurnNext)
	}
	// The skipping seat's own end step still ends its turn: it is not
	// skipped, only its next one.
	if got := hintAt(t, g, game.StepEnd, active, active); got != PassHintEndsTurn {
		t.Errorf("the active seat: pass_hint = %q, want %q", got, PassHintEndsTurn)
	}
}

func TestPassHintFollowsAnExtraTurn(t *testing.T) {
	g := passHintTable(t, 4)
	active := g.Turn.ActiveSeat
	normalNext := (active + 1) % 4
	extra := (active + 2) % 4
	g.WithWriteLock(func() { g.TakeExtraTurnsForEffect(g.Seats[extra].ID, uuid.Nil, 1) })

	// The extra turn comes first, so the seat that would normally be
	// next is not.
	if got := hintAt(t, g, game.StepEnd, normalNext, normalNext); got != "" {
		t.Errorf("the normal next seat before an extra turn: pass_hint = %q, want none", got)
	}
	if got := hintAt(t, g, game.StepEnd, extra, extra); got != PassHintYourTurnNext {
		t.Errorf("the extra turn's seat: pass_hint = %q, want %q", got, PassHintYourTurnNext)
	}
	// And the active player's pass still ends the turn.
	if got := hintAt(t, g, game.StepEnd, active, active); got != PassHintEndsTurn {
		t.Errorf("the active seat: pass_hint = %q, want %q", got, PassHintEndsTurn)
	}
}

func TestPassHintForYourOwnExtraTurnQueuedAfterThisOne(t *testing.T) {
	g := passHintTable(t, 2)
	active := g.Turn.ActiveSeat
	g.WithWriteLock(func() { g.TakeExtraTurnsForEffect(g.Seats[active].ID, uuid.Nil, 1) })
	// The pass still ends this turn, and the opponent is not next.
	if got := hintAt(t, g, game.StepEnd, active, active); got != PassHintEndsTurn {
		t.Errorf("active seat with an extra turn queued: pass_hint = %q, want %q", got, PassHintEndsTurn)
	}
	if got := hintAt(t, g, game.StepEnd, 1-active, 1-active); got != "" {
		t.Errorf("the opponent before my extra turn: pass_hint = %q, want none", got)
	}
}

func TestPassHintIsPerViewerAndNeverPublic(t *testing.T) {
	g := passHintTable(t, 2)
	active := g.Turn.ActiveSeat
	g.WithWriteLock(func() {
		g.Turn.Step = game.StepEnd
		g.Turn.PriorityHolder = active
	})
	v := ViewOfGame(g)
	if v.Turn.PassHint != "" {
		t.Errorf("the unfiltered view carries pass_hint %q", v.Turn.PassHint)
	}
	for _, viewer := range []string{"", SpectatorViewerID, g.Seats[1-active].ID.String()} {
		if got := FilterViewFor(v, viewer).Turn.PassHint; got != "" {
			t.Errorf("viewer %q sees pass_hint %q while seat %d holds priority", viewer, got, active)
		}
	}
	mine := FilterViewFor(v, g.Seats[active].ID.String())
	if mine.Turn.PassHint != PassHintEndsTurn {
		t.Fatalf("the active seat sees pass_hint %q, want %q", mine.Turn.PassHint, PassHintEndsTurn)
	}
	// Filtering a filtered view for someone else replaces the hint.
	if got := FilterViewFor(mine, g.Seats[1-active].ID.String()).Turn.PassHint; got != "" {
		t.Errorf("a second filter kept the first viewer's pass_hint %q", got)
	}
	b, err := json.Marshal(mine.Turn)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"pass_hint":"ends_turn"`) {
		t.Errorf("pass_hint missing from %s", b)
	}
	b, err = json.Marshal(TurnView{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "pass_hint") {
		t.Errorf("an unset pass_hint is on the wire: %s", b)
	}
}
