package game

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// move_tokens_test.go — #4522 (ADR 0178 amendment of 2026-11-05): a token
// creation that is part of a replaced move lands after the move, and the
// move's continuation waits for it, inline, across a CR 616 ordering
// prompt, across an undo and across a restore point.

// exileAndMakeTokenInstead is Kalitas, Traitor of Ghet's replacement as a
// catalog stub: "if a nontoken creature an opponent controls would die,
// instead exile that card and create a token" (a Goblin here).
func exileAndMakeTokenInstead() ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, g *Game, src *Card) bool {
			if src == nil || ev.Kind != RepEventMove || ev.OldZone != ZoneBattlefield || ev.NewZone != ZoneGraveyard {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.IsCreature() && !c.IsToken() && c.Controller != src.Controller
		},
		Replace: func(ev *ReplacementEvent, g *Game, src *Card) error {
			ev.NewZone = ZoneExile
			ev.NewZoneOwner = uuid.Nil
			return g.CreateTokensAfterReplacedMoveForEffect(ev, TokenCreation{
				Controller: src.Controller,
				Source:     src.InstanceID,
				Groups:     []TokenGroup{{Template: goblinToken(), Count: 1}},
			})
		},
		Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
		Label:      "exile instead, and create a token",
	}
}

// moveTokensBoard is a three-seat game where seat 0 controls the
// replacement and seat 1 a bear to lose. withPrompt gives seat 0 a doubler
// and a kind-swapper, two different creation replacements, so the creation
// asks CR 616.
func moveTokensBoard(t *testing.T, withPrompt bool) (*Game, *Player, uuid.UUID) {
	t.Helper()
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[0]
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"kalitas": {exileAndMakeTokenInstead()},
		"doubler": {tokenDoubler("double tokens")},
		"swapper": {tokenKindSwapper("a Clue instead")},
	})
	pushReplacementSource(g, "kalitas", me.ID)
	if withPrompt {
		pushReplacementSource(g, "doubler", me.ID)
		pushReplacementSource(g, "swapper", me.ID)
	}
	bear := pushBear(g, g.Seats[1].ID)
	return g, me, bear
}

// madeTokens counts the tokens the creation can have made.
func madeTokens(g *Game) int { return tokensNamed(g, "Goblin") + tokensNamed(g, "Clue") }

// destroyRecordingTokens records the token count each time the destroy's
// continuation runs.
func destroyRecordingTokens(t *testing.T, g *Game, ids []uuid.UUID) *[]int {
	t.Helper()
	var seen []int
	g.WithWriteLock(func() {
		err := g.DestroyPermanentsKeyedForEffect(ids, hookTail(func(g *Game, _ TailOutcome) error {
			seen = append(seen, madeTokens(g))
			return nil
		}))
		if err != nil {
			t.Fatalf("DestroyPermanentsKeyedForEffect: %v", err)
		}
	})
	return &seen
}

// Inline: the card is exiled, then the token is created, then the
// continuation runs and already sees the token.
func TestMoveTokensLandAfterTheMoveAndBeforeTheContinuation(t *testing.T) {
	g, me, bear := moveTokensBoard(t, false)
	var order []string
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventTokenCreated},
			AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
				if ev.Kind == RepEventCreateTokens {
					if g.Exile.Contains(bear) {
						order = append(order, "token after exile")
					} else {
						order = append(order, "token before exile")
					}
				}
				return false
			},
			Replace: func(*ReplacementEvent, *Game, *Card) error { return nil },
			Label:   "observer",
		})
	})
	seen := destroyRecordingTokens(t, g, []uuid.UUID{bear})
	if !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	if len(*seen) != 1 || (*seen)[0] != 1 {
		t.Fatalf("continuation saw tokens %v, want one run seeing 1 (the token lands first)", *seen)
	}
	if len(order) == 0 || order[0] != "token after exile" {
		t.Errorf("order = %v: the exile is the first instruction, the creation the second (CR 608.2c)", order)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.Controller != me.ID {
			t.Errorf("the token is controlled by %v, want the replacement's controller %v", c.Controller, me.ID)
		}
	}
	if len(g.movesAwaitingTokens) != 0 {
		t.Errorf("%d moves still parked after the creation", len(g.movesAwaitingTokens))
	}
	if len(g.PendingTriggers) != 0 {
		t.Errorf("%d triggers pending: the token is an instruction of the modified event, not a trigger", len(g.PendingTriggers))
	}
}

// Paused: the creation asks CR 616 and the destroy's continuation waits
// for the answer, then runs exactly once.
func TestMoveTokensContinuationWaitsForTheCreationPrompt(t *testing.T) {
	g, me, bear := moveTokensBoard(t, true)
	seen := destroyRecordingTokens(t, g, []uuid.UUID{bear})

	if !g.Exile.Contains(bear) {
		t.Fatal("the move lands before the creation's question is asked")
	}
	if len(*seen) != 0 {
		t.Fatalf("the continuation ran %d times with the creation prompt still open", len(*seen))
	}
	if n := madeTokens(g); n != 0 {
		t.Fatalf("%d tokens made before the prompt was answered", n)
	}
	onlyReplacementOrderPrompt(t, g, me.ID)
	answerOnlyPrompt(t, g)
	if len(*seen) != 1 || (*seen)[0] != 2 {
		t.Fatalf("after the answer the continuation saw %v, want one run seeing 2 tokens", *seen)
	}
	if len(g.movesAwaitingTokens) != 0 {
		t.Errorf("%d moves still parked after the answer", len(g.movesAwaitingTokens))
	}
}

// The token's ENTRY can pause too (two enters-tapped effects under an
// opponent), and the continuation waits for it the same way.
func TestMoveTokensContinuationWaitsForTheTokensEntryPrompt(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me, them := g.Seats[0], g.Seats[1]
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		"kalitas": {exileAndMakeTokenInstead()},
		"kismet":  {entersTappedForOthers("Kismet: enters tapped")},
		"thalia":  {entersTappedForOthers("Thalia: enters tapped")},
	})
	pushReplacementSource(g, "kalitas", me.ID)
	pushReplacementSource(g, "kismet", them.ID)
	pushReplacementSource(g, "thalia", them.ID)
	bear := pushBear(g, them.ID)
	seen := destroyRecordingTokens(t, g, []uuid.UUID{bear})

	if !g.Exile.Contains(bear) {
		t.Fatal("the creature was not exiled")
	}
	if len(*seen) != 0 {
		t.Fatal("the continuation ran with the token's entry prompt open")
	}
	onlyReplacementOrderPrompt(t, g, me.ID)
	answerOnlyPrompt(t, g)
	if len(*seen) != 1 || (*seen)[0] != 1 {
		t.Fatalf("after the answer the continuation saw %v, want one run seeing the token", *seen)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && !c.Tapped {
			t.Error("the token entered untapped")
		}
	}
}

// A wipe: the second leg waits for the first leg's creation prompt, and
// the continuation runs once after both.
func TestMoveTokensWipeSequencesEachLegsCreation(t *testing.T) {
	g, _, bear := moveTokensBoard(t, true)
	other := pushBear(g, g.Seats[2].ID)
	seen := destroyRecordingTokens(t, g, []uuid.UUID{bear, other})

	if g.Exile.Contains(other) {
		t.Fatal("the second leg moved with the first leg's creation prompt open")
	}
	answerOnlyPrompt(t, g)
	if len(*seen) != 0 {
		t.Fatal("the continuation ran with the second leg's creation prompt open")
	}
	answerOnlyPrompt(t, g)
	if len(*seen) != 1 || (*seen)[0] != 4 {
		t.Fatalf("after both answers: continuation runs %v; want one run seeing 4 tokens", *seen)
	}
	if !g.Exile.Contains(bear) || !g.Exile.Contains(other) {
		t.Error("both creatures are exiled")
	}
}

// A controller who has left the game gets no token (CR 800.4a), and the
// continuation is not stranded.
func TestMoveTokensSkipAControllerWhoLeft(t *testing.T) {
	g, me, bear := moveTokensBoard(t, false)
	g.WithWriteLock(func() {
		err := g.DestroyPermanentsKeyedForEffect(nil, KeyedTail{})
		if err != nil {
			t.Fatal(err)
		}
	})
	var seen *[]int
	g.WithWriteLock(func() {
		seq := moveLifeSequence{}
		me.Eliminated = true
		if err := g.continueMoveTokensLocked([]tokenMoveInstruction{{
			Controller: me.ID, Groups: []TokenGroup{{Template: goblinToken(), Count: 1}},
		}}, seq); err != nil {
			t.Fatal(err)
		}
	})
	if n := madeTokens(g); n != 0 {
		t.Fatalf("%d tokens for a controller who left", n)
	}
	g.WithWriteLock(func() { me.Eliminated = false })
	seen = destroyRecordingTokens(t, g, []uuid.UUID{bear})
	if len(*seen) != 1 {
		t.Fatalf("continuation ran %d times, want once", len(*seen))
	}
}

// An instruction declared before another replacement cancels the move
// still happens (CR 614.6), and the continuation runs once.
func TestMoveTokensSurviveALaterCancel(t *testing.T) {
	g, _, bear := moveTokensBoard(t, false)
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
	seen := destroyRecordingTokens(t, g, []uuid.UUID{bear})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the cancelled move moved the creature")
	}
	if len(*seen) != 1 || (*seen)[0] != 1 {
		t.Fatalf("continuation runs %v; want one run seeing the token", *seen)
	}
}

// A cost move never pauses (CR 601.2h / 602.2b): the creation that is
// part of it settles without asking, in the order gathered.
func TestMoveTokensOnACostSettleWithoutAsking(t *testing.T) {
	g, _, bear := moveTokensBoard(t, true)
	ran := 0
	g.WithWriteLock(func() {
		err := g.routeBattlefieldExitInBatchLocked(bear, zoneRoute{Sacrifice: true, MustSettleNow: true,
			batchLegs: &routeBatchRun{Leg: bear, Then: snapshotKeyedTail(hookTail(func(*Game, TailOutcome) error {
				ran++
				return nil
			}))}}, nil)
		if err != nil {
			t.Fatalf("sacrifice as a cost: %v", err)
		}
	})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts queued by a cost payment", len(g.PendingChoices))
	}
	if ran != 1 || madeTokens(g) != 2 {
		t.Fatalf("continuation ran %d times with %d tokens; want once and 2", ran, madeTokens(g))
	}
}

// Undo into the open creation prompt and answer again: the replay makes
// the same tokens and runs the continuation once, because the parked move
// is shared with the undo snapshot and never written.
func TestMoveTokensUndoAcrossTheCreationPromptReplaysTheSameWay(t *testing.T) {
	g, _, bear := moveTokensBoard(t, true)
	seen := destroyRecordingTokens(t, g, []uuid.UUID{bear})
	promptOpen := g.Clone()
	answerOnlyPrompt(t, g)
	if len(*seen) != 1 || madeTokens(g) != 2 {
		t.Fatalf("first run: continuation %v, %d tokens; want one run and 2", *seen, madeTokens(g))
	}
	*seen = nil
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if madeTokens(g) != 0 || len(g.movesAwaitingTokens) != 1 {
		t.Fatalf("the rewind left %d tokens and %d parked moves; want 0 and 1", madeTokens(g), len(g.movesAwaitingTokens))
	}
	answerOnlyPrompt(t, g)
	if len(*seen) != 1 || madeTokens(g) != 2 {
		t.Errorf("replay: continuation %v, %d tokens; want one run and 2", *seen, madeTokens(g))
	}
}

// A creation paused with the move parked behind it is a restore point:
// the carrier is written, and the restored answer makes the tokens and
// then runs the move's continuation once.
func TestMoveTokensPausedCreationIsARestorePoint(t *testing.T) {
	g, me, bear := moveTokensBoard(t, true)
	other := pushBear(g, g.Seats[2].ID)
	seen := destroyRecordingTokens(t, g, []uuid.UUID{bear, other})

	snap := throughJSON(t, g.CaptureSnapshot())
	if !snap.Restorable() {
		t.Fatalf("not a restore point: %+v", snap.Continuations)
	}
	if len(snap.MovesAwaitingTokens) != 1 || snap.MovesAwaitingTokens[0].ZoneRoute == nil {
		t.Fatalf("parked moves written: %+v; want one, with its route", snap.MovesAwaitingTokens)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	onlyReplacementOrderPrompt(t, restored, me.ID)
	answerOnlyPrompt(t, restored)
	if len(*seen) != 0 {
		t.Fatal("the restored table ran the continuation before the second leg")
	}
	if !restored.Exile.Contains(other) {
		t.Fatal("the restored answer did not go on to the second leg")
	}
	answerOnlyPrompt(t, restored)
	if len(*seen) != 1 || (*seen)[0] != 4 {
		t.Fatalf("restored: continuation runs %v; want one run seeing 4 tokens", *seen)
	}
	if len(restored.movesAwaitingTokens) != 0 {
		t.Errorf("%d moves still parked", len(restored.movesAwaitingTokens))
	}
}

// A move paused on a CR 616 question with a token creation declared is a
// restore point under its own key, and the restored answer still creates
// the token whichever way it is answered (CR 614.6: an instruction the
// move cannot carry out is ignored, the others still happen). A file
// that writes it under an older key is refused.
func TestAPausedMoveWithATokenCreationIsARestorePoint(t *testing.T) {
	for _, apply := range []bool{true, false} {
		g, cmdID := bouncedCommanderTable(t)
		owner := g.Seats[0]
		g.WithWriteLock(func() {
			f := g.PendingChoices[0].replacementResume
			f.ev.tokensAfterMove = []tokenMoveInstruction{{
				Source: cmdID, Controller: owner.ID,
				Groups: []TokenGroup{{Template: goblinToken(), Count: 1}},
			}}
		})
		snap := throughJSON(t, g.CaptureSnapshot())
		if !snap.Restorable() {
			t.Fatalf("not a restore point: %+v", snap.Continuations)
		}
		if got := snap.PendingChoices[0].PickThen; got != ReplacementResumeMoveTokens {
			t.Fatalf("frame written under %q, want %q", got, ReplacementResumeMoveTokens)
		}
		raw, err := json.Marshal(snap)
		if err != nil {
			t.Fatal(err)
		}
		var stale GameSnapshot
		if err := json.Unmarshal(raw, &stale); err != nil {
			t.Fatal(err)
		}
		stale.PendingChoices[0].PickThen = ReplacementResumeMove
		if _, err := stale.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
			t.Fatalf("a token creation under the plain move key: err = %v, want ErrUnknownEffectKey", err)
		}

		restored, err := snap.RestoreStrict()
		if err != nil {
			t.Fatalf("RestoreStrict: %v", err)
		}
		rowner := restored.Seats[0]
		prompt := expectCommanderPrompt(t, restored, rowner)
		if err := restored.ResolveOptionalReplacement(prompt.ID, rowner.ID, apply); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
		if apply {
			assertOnlyIn(t, cmdID, rowner.Command, rowner.Hand, restored.Battlefield)
		} else {
			assertOnlyIn(t, cmdID, rowner.Hand, rowner.Command, restored.Battlefield)
		}
		if n := tokensNamed(restored, "Goblin"); n != 1 {
			t.Errorf("apply=%v: %d Goblins, want 1", apply, n)
		}
	}
}

func TestCreateTokensAfterReplacedMoveRefusesMalformed(t *testing.T) {
	g := NewGame()
	who := uuid.New()
	grp := TokenGroup{Template: goblinToken(), Count: 1}
	for name, tc := range map[string]struct {
		ev   *ReplacementEvent
		spec TokenCreation
	}{
		"nil event":  {nil, TokenCreation{Controller: who, Groups: []TokenGroup{grp}}},
		"not a move": {&ReplacementEvent{Kind: RepEventDraw}, TokenCreation{Controller: who, Groups: []TokenGroup{grp}}},
		"attacking":  {&ReplacementEvent{Kind: RepEventMove}, TokenCreation{Controller: who, Attacking: who, Groups: []TokenGroup{grp}}},
		"together":   {&ReplacementEvent{Kind: RepEventMove}, TokenCreation{Controller: who, Together: true, Groups: []TokenGroup{grp}}},
		"blocking": {&ReplacementEvent{Kind: RepEventMove}, TokenCreation{Controller: who,
			Groups: []TokenGroup{{Template: goblinToken(), Count: 1, Blocking: who}}}},
	} {
		var err error
		g.WithWriteLock(func() { err = g.CreateTokensAfterReplacedMoveForEffect(tc.ev, tc.spec) })
		if !errors.Is(err, ErrInvalidParam) {
			t.Errorf("%s: err = %v, want ErrInvalidParam", name, err)
		}
	}
	ev := &ReplacementEvent{Kind: RepEventMove}
	g.WithWriteLock(func() {
		_ = g.CreateTokensAfterReplacedMoveForEffect(ev, TokenCreation{Groups: []TokenGroup{grp}})
		_ = g.CreateTokensAfterReplacedMoveForEffect(ev, TokenCreation{Controller: who, Groups: []TokenGroup{{Template: goblinToken()}}})
	})
	if len(ev.tokensAfterMove) != 0 {
		t.Errorf("a creation of nothing, or for nobody, declared %d instructions", len(ev.tokensAfterMove))
	}
}
