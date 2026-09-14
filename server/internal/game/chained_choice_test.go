package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// chained_choice_test.go — the composition contract, independent of any
// card: a prompt whose answer queues the next prompt, and what that
// costs a restore point.

// TestConfirmBranchQueuesTheNextPrompt is the mechanism in one test. The
// accept branch queues a second confirm; answering that one runs the
// tail. Nothing is asked before its predecessor is answered.
func TestConfirmBranchQueuesTheNextPrompt(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0].ID
	var order []string

	g.WithWriteLock(func() {
		g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser:  me,
			Question: "first?",
			OnAccept: func(g *Game) error {
				order = append(order, "first accepted")
				g.QueueConfirmForEffect(ConfirmPrompt{
					Chooser:  me,
					Question: "second?",
					OnAccept: func(*Game) error {
						order = append(order, "second accepted")
						return nil
					},
					OnDecline: func(*Game) error {
						order = append(order, "second declined")
						return nil
					},
				})
				return nil
			},
		})
	})

	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts queued, want 1", len(g.PendingChoices))
	}
	first := g.PendingChoices[0].ID
	if err := g.ResolveConfirm(first, me, true); err != nil {
		t.Fatalf("ResolveConfirm(first): %v", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatalf("%d prompts open after the first answer, want 1 (the second link)",
			len(g.PendingChoices))
	}
	second := g.PendingChoices[0]
	if second.ID == first {
		t.Fatal("the first prompt was not dequeued before its branch ran")
	}
	if second.Reason != "second?" {
		t.Errorf("second prompt reads %q", second.Reason)
	}
	if err := g.ResolveConfirm(second.ID, me, false); err != nil {
		t.Fatalf("ResolveConfirm(second): %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("%d prompts left open", len(g.PendingChoices))
	}
	want := []string{"first accepted", "second declined"}
	if len(order) != len(want) || order[0] != want[0] || order[1] != want[1] {
		t.Errorf("ran %v, want %v", order, want)
	}
}

// TestConfirmDeclineRunsTheOtherBranch — "no" is an answer, not a
// refusal to answer. Both the human path and the bot path reach it
// through the same {apply:false} payload.
func TestConfirmDeclineRunsTheOtherBranch(t *testing.T) {
	g := newRestorableGame(t)
	me := g.Seats[0].ID
	accepted, declined := false, false
	g.WithWriteLock(func() {
		g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser:   me,
			Question:  "either?",
			OnAccept:  func(*Game) error { accepted = true; return nil },
			OnDecline: func(*Game) error { declined = true; return nil },
		})
	})
	id := g.PendingChoices[0].ID
	if err := g.ResolveConfirm(id, me, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if accepted || !declined {
		t.Errorf("accepted=%v declined=%v, want the decline branch", accepted, declined)
	}
	if len(g.PendingChoices) != 0 {
		t.Error("a declined confirm stayed in the queue")
	}
}

// TestConfirmRejectsTheWrongChooser — an opponent cannot answer a
// prompt addressed to someone else, and the prompt survives the
// attempt.
func TestConfirmRejectsTheWrongChooser(t *testing.T) {
	g := newRestorableGame(t)
	me, them := g.Seats[0].ID, g.Seats[1].ID
	g.WithWriteLock(func() {
		g.QueueConfirmForEffect(ConfirmPrompt{Chooser: me, Question: "mine"})
	})
	id := g.PendingChoices[0].ID
	if err := g.ResolveConfirm(id, them, true); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("error = %v, want ErrNotTheChooser", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Error("a rejected answer dropped the prompt")
	}
}

// TestChooseCardsEnforcesItsBounds. The bounds live on the choice so
// the enumerator can offer exactly what the resolver accepts — the
// #544 lesson. This is the resolver half of that promise.
func TestChooseCardsEnforcesItsBounds(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	hand := handIDs(p, 3)
	var got []uuid.UUID
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  p.ID,
			Question: "pick two",
			Cards:    hand,
			Min:      2,
			Max:      2,
			Zone:     ZoneHand,
			Then: func(_ *Game, picked []uuid.UUID) error {
				got = picked
				return nil
			},
		})
	})
	id := g.PendingChoices[0].ID

	if err := g.ResolveChooseCards(id, p.ID, hand[:1]); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("one pick error = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveChooseCards(id, p.ID, hand); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("three picks error = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveChooseCards(id, p.ID, []uuid.UUID{hand[0], hand[0]}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("duplicate pick error = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveChooseCards(id, p.ID, []uuid.UUID{hand[0], uuid.New()}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("off-list pick error = %v, want ErrInvalidParam", err)
	}
	if len(g.PendingChoices) != 1 {
		t.Fatal("a rejected answer dropped the prompt; the client can no longer retry")
	}
	if err := g.ResolveChooseCards(id, p.ID, hand[:2]); err != nil {
		t.Fatalf("legal answer rejected: %v", err)
	}
	if len(got) != 2 || got[0] != hand[0] || got[1] != hand[1] {
		t.Errorf("continuation got %v, want %v", got, hand[:2])
	}
	if len(g.PendingChoices) != 0 {
		t.Error("an answered choose-cards stayed in the queue")
	}
}

// TestChooseCardsRechecksTheZone — the prompt is asynchronous and the
// board moves under it. A card that left the named zone between the
// question and the answer is not a legal pick any more.
func TestChooseCardsRechecksTheZone(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	hand := handIDs(p, 2)
	g.WithWriteLock(func() {
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser: p.ID,
			Cards:   hand,
			Min:     1,
			Max:     1,
			Zone:    ZoneHand,
			Then:    func(*Game, []uuid.UUID) error { return nil },
		})
		// The card is discarded while the prompt is open.
		if _, err := MoveCard(p.Hand, p.Graveyard, hand[0]); err != nil {
			t.Fatalf("MoveCard: %v", err)
		}
	})
	id := g.PendingChoices[0].ID
	if err := g.ResolveChooseCards(id, p.ID, hand[:1]); !errors.Is(err, ErrCardNotFound) {
		t.Errorf("error = %v, want ErrCardNotFound", err)
	}
	if err := g.ResolveChooseCards(id, p.ID, hand[1:]); err != nil {
		t.Errorf("the card still in hand was rejected: %v", err)
	}
}

// TestChainedChoiceFramesAreCensused is the restorability contract, and
// the reason it is a test: an uncounted continuation would let the
// server write a restore point that silently drops the rest of a card.
// Both new frames must make a snapshot non-restorable while open, and
// answering must give the restore point back.
func TestChainedChoiceFramesAreCensused(t *testing.T) {
	cases := []struct {
		name  string
		queue func(g *Game, me uuid.UUID)
	}{
		{"confirm", func(g *Game, me uuid.UUID) {
			g.QueueConfirmForEffect(ConfirmPrompt{
				Chooser:  me,
				Question: "?",
				OnAccept: func(*Game) error { return nil },
			})
		}},
		{"choose_cards", func(g *Game, me uuid.UUID) {
			g.QueueChooseCardsForEffect(ChooseCardsPrompt{
				Chooser: me,
				Cards:   handIDs(g.playerByIDLocked(me), 1),
				Then:    func(*Game, []uuid.UUID) error { return nil },
			})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newRestorableGame(t)
			me := g.Seats[0].ID
			if !g.CaptureSnapshot().Restorable() {
				t.Fatal("setup: baseline game is already not a restore point")
			}
			g.WithWriteLock(func() { tc.queue(g, me) })

			snap := g.CaptureSnapshot()
			if snap.Continuations.ChoiceResumeFrames != 1 {
				t.Errorf("census counted %d resume frames, want 1; census = %+v",
					snap.Continuations.ChoiceResumeFrames, snap.Continuations)
			}
			if snap.Restorable() {
				t.Fatal("a game holding a chained-choice continuation must not be a restore point")
			}
			if _, err := snap.RestoreStrict(); !errors.Is(err, ErrSnapshotNotRestorable) {
				t.Errorf("RestoreStrict error = %v, want ErrSnapshotNotRestorable", err)
			}

			// ... and answering it hands the restore point back. The
			// window is the length of the chain, not the rest of the
			// game.
			id := g.PendingChoices[0].ID
			var err error
			if tc.name == "confirm" {
				err = g.ResolveConfirm(id, me, true)
			} else {
				err = g.ResolveChooseCards(id, me, nil)
			}
			if err != nil {
				t.Fatalf("resolve: %v", err)
			}
			if !g.CaptureSnapshot().Restorable() {
				t.Error("the restore point did not come back after the chain finished")
			}
		})
	}
}

// TestChainedChoiceDataSurvivesASnapshot — the frames are dropped (and
// censused), but everything the PROMPT is made of has to round-trip, or
// a diagnostic restore renders a question with no options and no
// buttons.
func TestChainedChoiceDataSurvivesASnapshot(t *testing.T) {
	g := newRestorableGame(t)
	p := g.Seats[0]
	hand := handIDs(p, 2)
	g.WithWriteLock(func() {
		g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser:      p.ID,
			Question:     "pay 4 life?",
			AcceptLabel:  "Pay 4 life",
			DeclineLabel: "Put it on top",
			OnAccept:     func(*Game) error { return nil },
		})
		g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser: p.ID,
			Cards:   hand,
			Min:     1,
			Max:     2,
			Then:    func(*Game, []uuid.UUID) error { return nil },
		})
	})

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("lenient Restore: %v", err)
	}
	if len(restored.PendingChoices) != 2 {
		t.Fatalf("restored %d prompts, want 2", len(restored.PendingChoices))
	}
	c := restored.PendingChoices[0]
	if c.AcceptLabel != "Pay 4 life" || c.DeclineLabel != "Put it on top" {
		t.Errorf("confirm labels lost: %+v", c)
	}
	cc := restored.PendingChoices[1]
	if len(cc.ChooseCards) != 2 || cc.ChooseMin != 1 || cc.ChooseMax != 2 {
		t.Errorf("choose-cards prompt lost its candidates or bounds: %+v", cc)
	}
}

// handIDs returns the first n instance IDs in p's hand.
func handIDs(p *Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for _, c := range p.Hand.Cards {
		if len(out) == n {
			break
		}
		out = append(out, c.InstanceID)
	}
	return out
}
