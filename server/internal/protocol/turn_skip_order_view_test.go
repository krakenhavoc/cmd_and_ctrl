package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_skip_order_view_test.go — ADR 0188 Decision 12 (#3561): the
// which-skip question on the wire (PendingChoiceView.turn_skip_order,
// TurnView.between_turns) and its automatic answer's log line.

// pausedOnTheQuestion is a started game whose active seat owes a skip
// from a Magosi on the battlefield, has an extra turn queued, and whose
// extra turns a Trouble in Pairs-shaped replacement skips. The turn is
// passed, so the question is open.
func pausedOnTheQuestion(t *testing.T, rules map[string]game.AutoAnswer) (*game.Game, int) {
	t.Helper()
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	active := g.Turn.ActiveSeat
	me := g.Seats[active].ID
	magosi := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: magosi, Name: "Magosi, the Waterveil", TypeLine: "Land",
			Owner: me, Controller: me,
		})
		g.RegisterReplacementForTest(game.ReplacementEffect{
			Watches: []game.EventKind{game.EventExtraTurnBegin},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) bool {
				return ev.Kind == game.RepEventExtraTurn && ev.ExtraTurnSeat == active
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.Cancel()
				return nil
			},
			PureCancel: true,
			Label:      "Trouble in Pairs",
		})
		g.SkipNextTurnsForEffect(me, magosi, 1, "Magosi — skip your next turn")
		g.TakeExtraTurnsForEffect(me, uuid.Nil, 1)
	})
	if rules != nil {
		if err := g.SetAutoAnswers(me, rules); err != nil {
			t.Fatalf("SetAutoAnswers: %v", err)
		}
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	return g, active
}

func TestTheWhichSkipQuestionIsOnTheWireForEveryViewer(t *testing.T) {
	g, active := pausedOnTheQuestion(t, nil)
	v := ViewOfGame(g)
	if !v.Turn.BetweenTurns {
		t.Fatal("turn.between_turns is not set while the question is open")
	}
	for _, viewer := range g.Seats {
		fv := FilterViewFor(v, viewer.ID.String())
		var got *PendingChoiceView
		for i := range fv.PendingChoices {
			if fv.PendingChoices[i].Kind == string(game.PendingChoiceTurnSkipOrder) {
				got = &fv.PendingChoices[i]
			}
		}
		if got == nil || got.TurnSkipOrder == nil {
			t.Fatalf("viewer %s: no turn_skip_order prompt", viewer.Name)
		}
		o := got.TurnSkipOrder
		if got.Chooser != g.Seats[active].ID.String() || o.SkipSourceName != "Magosi, the Waterveil" ||
			len(o.Others) != 1 || o.Others[0] != "Trouble in Pairs" || o.Recommended != 0 {
			t.Errorf("viewer %s: %+v / %+v", viewer.Name, got, o)
		}
		if got.AcceptLabel != "Use my skip (Magosi, the Waterveil)" || got.DeclineLabel != "Let Trouble in Pairs skip it" {
			t.Errorf("viewer %s: labels %q / %q", viewer.Name, got.AcceptLabel, got.DeclineLabel)
		}
		if !fv.Turn.BetweenTurns {
			t.Errorf("viewer %s: between_turns missing", viewer.Name)
		}
	}
	// recommended: 0 is the value, so it is on the wire.
	b, err := json.Marshal(TurnSkipOrderView{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"recommended":0`) {
		t.Errorf("recommended 0 is missing from %s", b)
	}

	c := g.PendingChoices[0]
	if err := g.ResolveTurnSkipOrder(c.ID, c.Chooser, true); err != nil {
		t.Fatalf("ResolveTurnSkipOrder: %v", err)
	}
	if v := ViewOfGame(g); v.Turn.BetweenTurns {
		t.Error("between_turns is still set after the answer")
	}
}

func TestAnAutomaticWhichSkipAnswerIsLogged(t *testing.T) {
	for _, tc := range []struct {
		rule game.AutoAnswer
		text string
	}{
		{game.AutoAnswerAlways, "P%d used their own skip (Magosi, the Waterveil) on their extra turn (automatic)"},
		{game.AutoAnswerNever, "P%d let Trouble in Pairs skip their extra turn (automatic)"},
	} {
		t.Run(string(tc.rule), func(t *testing.T) {
			g, active := pausedOnTheQuestion(t, map[string]game.AutoAnswer{game.TurnSkipOrderAutoAnswerKey: tc.rule})
			id, _, ok := g.NextAutoAnswer()
			if !ok {
				t.Fatal("NextAutoAnswer offered nothing")
			}
			if _, err := g.AutoAnswer(id); err != nil {
				t.Fatalf("AutoAnswer: %v", err)
			}
			lines := logLinesOfKind(g, LogAutoAnswer)
			if len(lines) != 1 {
				t.Fatalf("%d auto_answer lines, want 1", len(lines))
			}
			want := strings.Replace(tc.text, "%d", string(rune('1'+active)), 1)
			if lines[0].Text != want {
				t.Errorf("text = %q, want %q", lines[0].Text, want)
			}
		})
	}
}
