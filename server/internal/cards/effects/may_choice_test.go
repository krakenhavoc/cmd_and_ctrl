package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_choice_test.go — #796. The free yes/no a resolving effect asks,
// tested as a primitive rather than only through the three cards that
// use it: the branches run once and in order, the prompt blocks the
// table, the question can be addressed to a seat other than the
// controller, and the whole thing rewinds.

// answerMayChoice answers the newest open MayChoice (a
// PendingChoiceConfirm) for `chooser`.
func answerMayChoice(t *testing.T, g *game.Game, chooser uuid.UUID, yes bool) {
	t.Helper()
	c := latestConfirmFor(g, chooser)
	if c == nil {
		t.Fatalf("no yes/no prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, chooser, yes); err != nil {
		t.Fatalf("ResolveConfirm(%v): %v", yes, err)
	}
}

// testMayYes / testMayNo / testMayOrder are what the keyed probe answer
// (testMayProbe) records (#4253 retired the OnYes / OnNo closures): how
// many times each branch ran, and the order things happened in.
var (
	testMayYes, testMayNo int
	testMayOrder          []string
	testMayProbe          = MayYesNo("test/4253-may-choice-probe",
		func(*Context, game.ConfirmAnswer) error {
			testMayYes++
			testMayOrder = append(testMayOrder, "yes")
			return nil
		},
		func(*Context, game.ConfirmAnswer) error {
			testMayNo++
			testMayOrder = append(testMayOrder, "no")
			return nil
		})
)

// resetMayProbe zeroes testMayProbe's record.
func resetMayProbe() { testMayYes, testMayNo, testMayOrder = 0, 0, nil }

// latestConfirmFor returns the newest open confirm prompt owed by a
// seat, or nil.
func latestConfirmFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	var out *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

// TestMayChoiceRunsExactlyOneBranchInOrder is the primitive's core
// contract: the yes branch runs on yes, the no branch on no, each
// exactly once, and neither runs before the answer arrives.
func TestMayChoiceRunsExactlyOneBranchInOrder(t *testing.T) {
	for _, tc := range []struct {
		name            string
		answer          bool
		wantYes, wantNo int
	}{
		{"yes", true, 1, 0},
		{"no", false, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			resetMayProbe()
			g.WithWriteLock(func() {
				ctx := NewContext(g, &game.StackItem{Controller: me.ID})
				testMayOrder = append(testMayOrder, "before")
				if err := (MayChoice{
					Question: "test — do the thing?",
					Then:     testMayProbe,
				}).Apply(ctx); err != nil {
					t.Fatalf("Apply: %v", err)
				}
			})
			if testMayYes != 0 || testMayNo != 0 {
				t.Fatal("no branch runs until the prompt is answered")
			}
			if len(g.PendingChoices) != 1 || g.PendingChoices[0].Chooser != me.ID {
				t.Fatalf("the prompt is addressed to the controller: %+v", g.PendingChoices)
			}
			answerMayChoice(t, g, me.ID, tc.answer)
			if testMayYes != tc.wantYes || testMayNo != tc.wantNo {
				t.Errorf("yes ran %d (want %d), no ran %d (want %d)", testMayYes, tc.wantYes, testMayNo, tc.wantNo)
			}
			if order := testMayOrder; len(order) != 2 || order[0] != "before" {
				t.Errorf("the branch runs after the rest of the effect, not before it: %v", order)
			}
			if len(g.PendingChoices) != 0 {
				t.Errorf("the prompt is dequeued before its branch runs: %+v", g.PendingChoices)
			}
		})
	}
}

// TestMayChoiceCanBeAddressedToAnotherSeat is the half #568 builds on:
// Player names any seat, and only that seat may answer.
func TestMayChoiceCanBeAddressedToAnotherSeat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	resetMayProbe()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (MayChoice{
			Player:   opp.ID,
			Question: "test — will you?",
			Then:     testMayProbe,
		}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	c := latestConfirmFor(g, opp.ID)
	if c == nil {
		t.Fatalf("the prompt is addressed to the opponent: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err == nil {
		t.Error("only the chooser may answer")
	}
	if testMayYes != 0 || testMayNo != 0 {
		t.Fatal("a refused answer runs no branch")
	}
	answerMayChoice(t, g, opp.ID, true)
	if testMayYes != 1 {
		t.Error("the opponent's yes runs the branch")
	}
}

// TestMayChoiceBlocksTheTableUntilAnswered — the prompt is gated like
// every other resolution-time decision (#791): the cursor does not
// walk past a question whose answer is the rest of the card.
func TestMayChoiceBlocksTheTableUntilAnswered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (MayChoice{Question: "test — do the thing?", Then: testMayProbe}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	if !game.ChoiceBlocksTable(g.PendingChoices[0].Kind) {
		t.Fatal("a resolution-time yes/no blocks the table")
	}
}

// TestMayChoiceUndoesAcrossThePrompt is the undo assertion the issue
// asks for: rewinding past the answer puts the question back, and
// answering the restored prompt lands the same way — so the branch is
// not holding a pointer into the game that queued it.
func TestMayChoiceUndoesAcrossThePrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	source := uuid.New()
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: source})
		if err := (MayChoice{
			Question: "test — lose 3 life?",
			Then:     mayChoiceTestLoseThree,
		}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	// Undo is Clone + RestoreFrom, the way the engine does it.
	snap := g.Clone()
	answerMayChoice(t, g, me.ID, true)
	if g.Seats[0].Life != life-3 {
		t.Fatalf("the yes branch ran: life %d → %d", life, g.Seats[0].Life)
	}
	g.RestoreFrom(snap)
	if g.Seats[0].Life != life {
		t.Errorf("undo rewinds the branch: life %d, want %d", g.Seats[0].Life, life)
	}
	c := latestConfirmFor(g, me.ID)
	if c == nil {
		t.Fatalf("undo puts the question back: %+v", g.PendingChoices)
	}
	// The restored prompt is answerable and lands the same way. That
	// is the whole undo-safety claim: the branch resolves against the
	// game it is handed, which after an undo is the restored one.
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm after undo: %v", err)
	}
	if g.Seats[0].Life != life-3 {
		t.Errorf("answering the restored prompt: life %d, want %d", g.Seats[0].Life, life-3)
	}
}

// mayChoiceTestLoseThree is a registered yes — the shape the primitive's
// doc comment asks card files for, and the one that makes the undo test
// mean something: it captures no game and no player pointer, only what it
// reads off the Context it is handed.
var mayChoiceTestLoseThree = MayYes("test/4253-may-choice-lose-three", func(ctx *Context, _ game.ConfirmAnswer) error {
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -3)
})
