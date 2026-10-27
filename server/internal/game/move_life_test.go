package game

import (
	"testing"

	"github.com/google/uuid"
)

// move_life_test.go — ADR 0178 (#2950): a life change that is part of a
// replaced move lands after the move, and the move's continuation waits
// for it, inline and across a CR 616 ordering prompt.

// exileAndGainInstead is The Darkness Crystal's replacement as a test
// fixture: "if a creature controlled by someone other than gainer would
// die, instead exile it and gainer gains `gain` life".
func exileAndGainInstead(gainer uuid.UUID, gain int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
			if ev.Kind != RepEventMove || ev.OldZone != ZoneBattlefield || ev.NewZone != ZoneGraveyard {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.IsCreature() && c.Controller != gainer
		},
		Replace: func(ev *ReplacementEvent, g *Game, _ *Card) error {
			ev.NewZone = ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return g.ChangeLifeAfterReplacedMoveForEffect(ev, uuid.Nil, gainer, gain)
		},
		Label: "exile instead, and gain life",
	}
}

// moveLifeBoard is a three-seat game where seat 0 gains from the
// replacement and owns a bear for seat 1 to lose. withPrompt adds the two
// different life replacements on seat 0 that make its gain ask CR 616.
func moveLifeBoard(t *testing.T, withPrompt bool) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[0]
	bear := pushBear(g, g.Seats[1].ID)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(exileAndGainInstead(me.ID, 2))
		if withPrompt {
			g.RegisterReplacementForTest(plusOneLifeReplacement(me.ID))
			g.RegisterReplacementForTest(doubleLifeReplacement(me.ID))
		}
	})
	return g, me, bear
}

// destroyRecordingLife records the gainer's life each time the destroy's
// continuation runs.
func destroyRecordingLife(t *testing.T, g *Game, me *Player, ids []uuid.UUID) *[]int {
	t.Helper()
	var seen []int
	g.WithWriteLock(func() {
		err := g.DestroyPermanentsKeyedForEffect(ids, hookTail(func(_ *Game, o TailOutcome) error {
			seen = append(seen, me.Life)
			return nil
		}))
		if err != nil {
			t.Fatalf("DestroyPermanentsKeyedForEffect: %v", err)
		}
	})
	return &seen
}

// Inline: the card is exiled, the gain lands, then the continuation
// runs and already sees the gain.
func TestMoveLifeLandsAfterTheMoveAndBeforeTheContinuation(t *testing.T) {
	g, me, bear := moveLifeBoard(t, false)
	life := me.Life
	var order []string
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
				if ev.Kind == RepEventLife && ev.LifePlayer == me.ID {
					if g.Exile.Contains(bear) {
						order = append(order, "gain after exile")
					} else {
						order = append(order, "gain before exile")
					}
				}
				return false
			},
			Replace: func(*ReplacementEvent, *Game, *Card) error { return nil },
			Label:   "observer",
		})
	})
	seen := destroyRecordingLife(t, g, me, []uuid.UUID{bear})
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	if len(*seen) != 1 || (*seen)[0] != life+2 {
		t.Fatalf("continuation saw lives %v, want one run at %d (the gain lands first)", *seen, life+2)
	}
	if len(order) == 0 || order[0] != "gain after exile" {
		t.Errorf("order = %v: the exile is the first instruction, the gain the second (CR 608.2c)", order)
	}
}

// Paused: the gain asks CR 616 and the destroy's continuation waits for
// the answer, then runs exactly once.
func TestMoveLifeContinuationWaitsForTheLifePrompt(t *testing.T) {
	g, me, bear := moveLifeBoard(t, true)
	life := me.Life
	seen := destroyRecordingLife(t, g, me, []uuid.UUID{bear})

	if !g.Exile.Contains(bear) {
		t.Fatal("the move lands before the life question is asked")
	}
	if len(*seen) != 0 {
		t.Fatalf("the continuation ran %d times with the life prompt still open", len(*seen))
	}
	if me.Life != life {
		t.Fatalf("life moved to %d before the prompt was answered", me.Life)
	}
	answerOnlyPrompt(t, g)
	if len(*seen) != 1 {
		t.Fatalf("the continuation ran %d times after the answer, want once", len(*seen))
	}
	// Offered order: one more, then double: (2+1)*2.
	if (*seen)[0] != life+6 || me.Life != life+6 {
		t.Errorf("life %d, continuation saw %v; want %d", me.Life, *seen, life+6)
	}
}

// A wipe: the second leg waits for the first leg's life prompt, and the
// continuation runs once after both.
func TestMoveLifeWipeSequencesEachLegsGain(t *testing.T) {
	g, me, bear := moveLifeBoard(t, true)
	other := pushBear(g, g.Seats[2].ID)
	life := me.Life
	seen := destroyRecordingLife(t, g, me, []uuid.UUID{bear, other})

	answerOnlyPrompt(t, g)
	if len(*seen) != 0 {
		t.Fatal("the continuation ran with the second leg's life prompt open")
	}
	answerOnlyPrompt(t, g)
	if len(*seen) != 1 || me.Life != life+12 {
		t.Fatalf("after both answers: continuation runs %v, life %d; want one run and %d", *seen, me.Life, life+12)
	}
	if !g.Exile.Contains(bear) || !g.Exile.Contains(other) {
		t.Error("both creatures are exiled")
	}
}

// A gain that can't happen (CR 119.7) or is replaced away still lets the
// move's continuation run.
func TestMoveLifeCancelledGainStillRunsTheContinuation(t *testing.T) {
	g, me, bear := moveLifeBoard(t, false)
	life := me.Life
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventChangeLife},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventLife && ev.LifePlayer == me.ID
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Canceled = true
				return nil
			},
			PureCancel: true,
			Label:      "your life total can't change",
		})
	})
	seen := destroyRecordingLife(t, g, me, []uuid.UUID{bear})
	if len(*seen) != 1 || me.Life != life {
		t.Fatalf("continuation runs %v, life %d; want one run and no gain", *seen, me.Life)
	}
}

// A gainer who has left the game gains nothing, and the continuation is
// not stranded.
func TestMoveLifeSkipsAGainerWhoLeft(t *testing.T) {
	g, me, bear := moveLifeBoard(t, false)
	g.WithWriteLock(func() { me.Eliminated = true })
	seen := destroyRecordingLife(t, g, me, []uuid.UUID{bear})
	if len(*seen) != 1 {
		t.Fatalf("continuation ran %d times, want once", len(*seen))
	}
}

// A cost move never pauses (CR 601.2h / 602.2b): the life change that is
// part of it settles without asking, in the order gathered.
func TestMoveLifeOnACostSettlesWithoutAsking(t *testing.T) {
	g, me, bear := moveLifeBoard(t, true)
	life := me.Life
	ran := 0
	g.WithWriteLock(func() {
		err := g.routeBattlefieldExitInBatchThenLocked(bear, zoneRoute{Sacrifice: true, MustSettleNow: true}, nil, func(*Game) error {
			ran++
			return nil
		})
		if err != nil {
			t.Fatalf("sacrifice as a cost: %v", err)
		}
	})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts queued by a cost payment", len(g.PendingChoices))
	}
	if ran != 1 || me.Life <= life {
		t.Fatalf("continuation ran %d times, life %d -> %d; want once and a gain", ran, life, me.Life)
	}
}

// An instruction declared before another replacement cancels the move
// still happens (CR 614.6), and the continuation runs once.
func TestMoveLifeSurvivesALaterCancel(t *testing.T) {
	g, me, bear := moveLifeBoard(t, false)
	life := me.Life
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == bear && ev.NewZone == ZoneExile
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Canceled = true
				return nil
			},
			PureCancel: true,
			Label:      "it stays",
		})
	})
	seen := destroyRecordingLife(t, g, me, []uuid.UUID{bear})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the cancelled move moved the creature")
	}
	if len(*seen) != 1 || me.Life != life+2 {
		t.Fatalf("continuation runs %v, life %d; want one run and +2", *seen, me.Life)
	}
}

// Undo into the open life prompt and answer again: the replay gains the
// same and runs the continuation again, because the snapshot has its own
// copy of the move's route.
func TestMoveLifeUndoIntoTheLifePromptReplaysTheSameWay(t *testing.T) {
	g, me, bear := moveLifeBoard(t, true)
	life := me.Life
	seen := destroyRecordingLife(t, g, me, []uuid.UUID{bear})
	promptOpen := g.Clone()

	answerOnlyPrompt(t, g)
	if len(*seen) != 1 || me.Life != life+6 {
		t.Fatalf("first answer: runs %v, life %d", *seen, me.Life)
	}

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	me = g.Seats[0]
	if me.Life != life || !g.Exile.Contains(bear) {
		t.Fatalf("after rewinding into the prompt: life %d, exiled %v", me.Life, g.Exile.Contains(bear))
	}
	answerOnlyPrompt(t, g)
	if len(*seen) != 2 {
		t.Fatalf("the replayed answer ran the continuation %d times in all, want 2 (once per answer)", len(*seen))
	}
	if me.Life != life+6 {
		t.Errorf("replayed life = %d, want %d", me.Life, life+6)
	}
}

// The move itself pauses first (two different replacements on the
// move), and the life prompt follows only once that is answered.
func TestMoveLifeAfterAPausedMove(t *testing.T) {
	g, me, bear := moveLifeBoard(t, true)
	life := me.Life
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
				if ev.Kind != RepEventMove || ev.CardID != bear || ev.OldZone != ZoneBattlefield {
					return false
				}
				return ev.NewZone == ZoneGraveyard
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.ExileCounter = "test"
				return nil
			},
			Label: "a second move replacement",
		})
	})
	seen := destroyRecordingLife(t, g, me, []uuid.UUID{bear})
	if len(g.PendingChoices) != 1 || g.PendingChoices[0].Kind != PendingChoiceReplacementOrder {
		t.Fatalf("want the move's own CR 616 prompt first, have %d choices", len(g.PendingChoices))
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the paused move moved the creature")
	}
	// Either order ends in exile: the second replacement no longer
	// applies once the destination is exile, and the exile still applies
	// after it.
	answerOnlyPrompt(t, g)
	if !g.Exile.Contains(bear) {
		t.Fatal("the answered move did not exile the creature")
	}
	if len(*seen) != 0 {
		t.Fatal("the continuation ran with the life prompt still open")
	}
	if me.Life != life {
		t.Fatalf("life moved to %d before its own prompt was answered", me.Life)
	}
	answerOnlyPrompt(t, g)
	if len(*seen) != 1 || me.Life != life+6 {
		t.Fatalf("runs %v, life %d; want one run and %d", *seen, me.Life, life+6)
	}
}
