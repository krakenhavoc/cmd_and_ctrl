package game

import (
	"testing"

	"github.com/google/uuid"
)

// sacrifice_run_test.go pins #1019 and #3664: a prompted sacrifice has
// a continuation, and it runs once, after the LAST asked seat has
// settled.
//
// The fan-out ("each player sacrifices", EachPlayerSacrificesThenForEffect)
// is ADR 0175's ordered collector since #3664: one seat is asked at a
// time in APNAP order, nothing moves until the last has chosen, and
// every pick then leaves in one simultaneous exit (CR 101.4, 608.2e).
// A seat with exactly as many candidates as it owes is collected
// without a prompt, so the tests below give each asked seat two bears.
//
// What a run has to survive, one test each below: several seats
// answering, a leg an optional replacement has PAUSED, a seat that
// leaves mid-prompt, a seat that had nothing to sacrifice, a seat
// whose candidates leave before its turn, and an undo across a
// half-answered fan-out.

// sacrificeRunCall records one run of a continuation.
type sacrificeRunCall struct {
	ran int
	got PromptedSacrifices
}

// eachPlayerSacrificesThen starts a fan-out run over every seat and
// records what its continuation is handed.
func eachPlayerSacrificesThen(t *testing.T, g *Game, source, except uuid.UUID) *sacrificeRunCall {
	t.Helper()
	out := &sacrificeRunCall{}
	g.WithWriteLock(func() {
		err := g.EachPlayerSacrificesThenForEffect(source, except, nil, "sacrifice a permanent",
			func(_ *Game, sacrificed PromptedSacrifices) error {
				out.ran++
				out.got = sacrificed
				return nil
			})
		if err != nil {
			t.Fatalf("EachPlayerSacrificesThenForEffect: %v", err)
		}
	})
	return out
}

// isSacrificePromptKind reports whether c asks its chooser to
// sacrifice: the single-seat sacrifice_choice, or a fan-out's
// choose_cards leg (ChooseSacrifice).
func isSacrificePromptKind(c *PendingChoice) bool {
	return c != nil && (c.Kind == PendingChoiceSacrifice ||
		(c.Kind == PendingChoiceChooseCards && c.ChooseDestination == ChooseSacrifice))
}

// sacrificePromptFor returns the open sacrifice prompt owed by `seat`,
// or nil.
func sacrificePromptFor(g *Game, seat uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if isSacrificePromptKind(c) && c.Chooser == seat {
			return c
		}
	}
	return nil
}

// answerSacrifice answers the open sacrifice prompt owed by `seat`
// with `card`.
func answerSacrifice(t *testing.T, g *Game, seat, card uuid.UUID) {
	t.Helper()
	c := sacrificePromptFor(g, seat)
	if c == nil {
		t.Fatalf("seat %s owes no sacrifice prompt", seat)
	}
	if c.Kind == PendingChoiceChooseCards {
		if err := g.ResolveChooseCards(c.ID, seat, []uuid.UUID{card}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		return
	}
	if err := g.ResolveSacrificeChoice(c.ID, seat, card); err != nil {
		t.Fatalf("ResolveSacrificeChoice: %v", err)
	}
}

// openSacrificePrompts counts the sacrifice prompts still queued.
func openSacrificePrompts(g *Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if isSacrificePromptKind(c) {
			n++
		}
	}
	return n
}

// twoBears gives `seat` two bears, so a one-permanent edict has a real
// choice to ask it, and returns the first.
func twoBears(g *Game, seat uuid.UUID) uuid.UUID {
	first := pushBear(g, seat)
	pushBear(g, seat)
	return first
}

// TestAPromptedSacrificeRunWaitsForEveryAskedSeat is the headline, and
// Rise of the Witch-king's shape: three seats are asked, and the
// continuation runs exactly once — after the third answer, not after
// the third QUESTION.
//
// #3664: the seats are asked one at a time in APNAP order, a later
// seat cannot answer first, and no picked permanent moves until the
// last seat has chosen.
func TestAPromptedSacrificeRunWaitsForEveryAskedSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	seats := g.Seats[:3]
	source := departureTestSource(g, seats[0].ID, "Rise of the Witch-king")
	bears := map[uuid.UUID]uuid.UUID{}
	for _, p := range seats {
		bears[p.ID] = twoBears(g, p.ID)
	}

	run := eachPlayerSacrificesThen(t, g, source, g.Seats[3].ID)

	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with prompts open, want 0 — "+
			"queueing a question is not sacrificing anything", run.ran)
	}
	for i, p := range seats {
		if openSacrificePrompts(g) != 1 {
			t.Fatalf("before seat %d answers: %d prompts open, want exactly 1 (CR 101.4: "+
				"one player chooses at a time)", i, openSacrificePrompts(g))
		}
		if sacrificePromptFor(g, p.ID) == nil {
			t.Fatalf("seat %d (%s) is not the one being asked; APNAP order is broken", i, p.Name)
		}
		for _, later := range seats[i+1:] {
			if sacrificePromptFor(g, later.ID) != nil {
				t.Fatalf("seat %s can answer before %s", later.Name, p.Name)
			}
		}
		for _, q := range seats {
			if findBattlefieldCard(g, bears[q.ID]) == nil {
				t.Fatalf("%s's pick left the battlefield before every player had chosen (CR 608.2e)", q.Name)
			}
		}
		answerSacrifice(t, g, p.ID, bears[p.ID])
	}

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1", run.ran)
	}
	if run.got.Count() != 3 {
		t.Errorf("the run sacrificed %d permanents, want 3 (%v)", run.got.Count(), run.got)
	}
	for _, p := range seats {
		if !run.got.Sacrificed(p.ID) {
			t.Errorf("seat %s answered and is reported as having sacrificed nothing", p.Name)
		}
		if got := run.got.By(p.ID); len(got) != 1 || got[0] != bears[p.ID] {
			t.Errorf("seat %s sacrificed %v, want the bear it named", p.Name, got)
		}
	}
	for i, e := range run.got {
		if e.Seat != seats[i].ID {
			t.Errorf("entry %d is seat %s, want the ask order (%s)", i, e.Seat, seats[i].Name)
		}
	}
}

// TestAPromptedSacrificeRunWaitsForAPausedLeg is the paused-leg
// half: a sacrificed card whose "may" replacement is pending (a test
// one; a commander was the vehicle before ADR 0115) is STILL ON THE
// BATTLEFIELD while its owner is asked, so a run that settled on the
// line after the answer would pay out with the permanent still in
// play.
func TestAPromptedSacrificeRunWaitsForAPausedLeg(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Rise of the Witch-king")
	commander := seatDetouredCard(t, g, g.Battlefield, me)

	run := eachPlayerSacrificesThen(t, g, source, g.Seats[1].ID)
	answerSacrifice(t, g, me.ID, commander)

	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0", run.ran)
	}
	if findBattlefieldCard(g, commander) == nil {
		t.Error("a paused leg has moved nothing")
	}

	prompt := expectCommanderPrompt(t, g, me)
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times after the CR 903.9 answer, want 1", run.ran)
	}
	if !run.got.Sacrificed(me.ID) {
		t.Errorf("the commander took the command zone and the run reports nothing sacrificed: "+
			"CR 701.17a's sacrifice is the move OFF the battlefield, and only where the card "+
			"went was replaced (%v)", run.got)
	}
	if !me.Command.Contains(commander) {
		t.Error("the commander is in the command zone")
	}
}

// TestAPromptedSacrificeRunIgnoresASeatWithNothingToSacrifice: a seat
// with no legal permanent is skipped (CR 701.21a's "if you can"), so
// it is never part of the run and the run does not wait for it.
func TestAPromptedSacrificeRunIgnoresASeatWithNothingToSacrifice(t *testing.T) {
	g := newActiveGame(t)
	me, empty := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Fleshbag Marauder")
	bear := twoBears(g, me.ID)

	run := eachPlayerSacrificesThen(t, g, source, uuid.Nil)

	if openSacrificePrompts(g) != 1 {
		t.Fatalf("expected one prompt — the other seat controls nothing — got %d",
			openSacrificePrompts(g))
	}
	answerSacrifice(t, g, me.ID, bear)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", run.ran)
	}
	if len(run.got) != 1 || run.got[0].Seat != me.ID {
		t.Errorf("the run reports %v, want one entry for the seat that was ASKED", run.got)
	}
	if run.got.Sacrificed(empty.ID) {
		t.Error("a seat that was never asked is reported as having sacrificed something")
	}
}

// TestAPromptedSacrificeRunWithNobodyToAskStillRunsTheRestOfTheCard is
// the #544 rule at queue time: a continuation is the rest of a card
// that is paused mid-resolution, so an instruction that could not ask
// anybody still finishes.
func TestAPromptedSacrificeRunWithNobodyToAskStillRunsTheRestOfTheCard(t *testing.T) {
	g := newActiveGame(t)
	source := departureTestSource(g, g.Seats[0].ID, "Fleshbag Marauder")
	g.WithWriteLock(func() { g.Battlefield.Cards = nil })

	run := eachPlayerSacrificesThen(t, g, source, uuid.Nil)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times with nobody to ask, want 1 — a card that asks "+
			"a question nobody can answer does not stop halfway", run.ran)
	}
	if run.got.Count() != 0 {
		t.Errorf("nothing was sacrificed, the run reports %v", run.got)
	}
}

// TestASeatWhoseCandidatesLeaveBeforeItsTurnIsSkipped: the later seat's
// permanents leave by some other route while an earlier seat is still
// choosing, so when its turn comes it has nothing to choose and the
// run finishes rather than waiting for an answer that can never come.
func TestASeatWhoseCandidatesLeaveBeforeItsTurnIsSkipped(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Fleshbag Marauder")
	mine := twoBears(g, me.ID)
	theirs := []uuid.UUID{pushBear(g, them.ID), pushBear(g, them.ID)}

	run := eachPlayerSacrificesThen(t, g, source, uuid.Nil)
	if sacrificePromptFor(g, me.ID) == nil {
		t.Fatal("the active player is asked first")
	}

	g.WithWriteLock(func() {
		for _, id := range theirs {
			if err := g.ExileCardForEffect(id); err != nil {
				t.Fatalf("ExileCardForEffect: %v", err)
			}
		}
	})
	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with a prompt still open, want 0", run.ran)
	}

	answerSacrifice(t, g, me.ID, mine)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — a seat with nothing left must be "+
			"skipped, or the run waits for an answer that is never coming", run.ran)
	}
	if !run.got.Sacrificed(me.ID) {
		t.Error("the seat that answered sacrificed its bear")
	}
	if got := run.got.By(them.ID); got != nil {
		t.Errorf("the emptied seat reports %v, want nothing", got)
	}
}

// TestASeatLeavingMidPromptSettlesItsSacrificeLeg is the departure
// sweep's half: the seat being asked concedes, its leg drops with no
// picks, and the rest of the instruction finishes for everybody else —
// including the earlier seat's pick, which had not moved yet.
func TestASeatLeavingMidPromptSettlesItsSacrificeLeg(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Rise of the Witch-king")
	mine := twoBears(g, caster.ID)
	twoBears(g, leaver.ID)

	run := eachPlayerSacrificesThen(t, g, source, uuid.Nil)
	answerSacrifice(t, g, caster.ID, mine)
	if sacrificePromptFor(g, leaver.ID) == nil {
		t.Fatal("the second seat is asked after the first answers")
	}
	if findBattlefieldCard(g, mine) == nil {
		t.Fatal("the caster's pick moved before the second seat had chosen")
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — a seat that leaves mid-prompt must "+
			"not strand the run", run.ran)
	}
	if !run.got.Sacrificed(caster.ID) {
		t.Error("the surviving seat sacrificed its bear")
	}
	if run.got.Sacrificed(leaver.ID) {
		t.Errorf("the departed seat is reported as having sacrificed something: %v", run.got)
	}
}

// TestAnUndoAcrossAHalfAnsweredSacrificeRunReplaysIdentically is the
// clone contract: an undo that rewound the queue and not the collected
// picks would replay the second answer against a stale receipt list and
// sacrifice the wrong set.
func TestAnUndoAcrossAHalfAnsweredSacrificeRunReplaysIdentically(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Fleshbag Marauder")
	mine := twoBears(g, me.ID)
	theirs := twoBears(g, them.ID)

	run := eachPlayerSacrificesThen(t, g, source, uuid.Nil)

	var snapshot *Game
	g.WithWriteLock(func() { snapshot = g.cloneLocked() })

	answerSacrifice(t, g, me.ID, mine)
	if run.ran != 0 {
		t.Fatalf("one answer of two ran the continuation %d times, want 0", run.ran)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
	if sacrificePromptFor(g, me.ID) == nil || openSacrificePrompts(g) != 1 {
		t.Fatalf("the undo restored %d prompts, want the first seat's one", openSacrificePrompts(g))
	}
	if run.ran != 0 {
		t.Fatalf("the undo itself ran the continuation %d times", run.ran)
	}

	answerSacrifice(t, g, me.ID, mine)
	if run.ran != 0 {
		t.Fatalf("the replayed answer ran the continuation %d times, want 0", run.ran)
	}
	answerSacrifice(t, g, them.ID, theirs)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1", run.ran)
	}
	if run.got.Count() != 2 {
		t.Errorf("the run sacrificed %d permanents, want 2 (%v)", run.got.Count(), run.got)
	}
	if got := run.got.By(me.ID); len(got) != 1 {
		t.Errorf("the replayed seat sacrificed %v, want exactly one permanent", got)
	}
}

// TestPlayerSacrificesThenCountsTheWholeInstruction: "sacrifice two
// lands" is ONE run of two prompts, not two runs — a caller looping
// the entry point would start two runs and pay out twice.
func TestPlayerSacrificesThenCountsTheWholeInstruction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Lotus Field")
	first := pushBear(g, me.ID)
	second := pushBear(g, me.ID)

	out := &sacrificeRunCall{}
	g.WithWriteLock(func() {
		err := g.PlayerSacrificesThenForEffect(source, me.ID, nil, "sacrifice a permanent", 2,
			func(_ *Game, sacrificed PromptedSacrifices) error {
				out.ran++
				out.got = sacrificed
				return nil
			})
		if err != nil {
			t.Fatalf("PlayerSacrificesThenForEffect: %v", err)
		}
	})

	if openSacrificePrompts(g) != 2 {
		t.Fatalf("expected two prompts for one instruction, got %d", openSacrificePrompts(g))
	}
	answerSacrifice(t, g, me.ID, first)
	if out.ran != 0 {
		t.Fatalf("the continuation ran after the first of two answers")
	}
	answerSacrifice(t, g, me.ID, second)

	if out.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", out.ran)
	}
	if out.got.Count() != 2 {
		t.Errorf("the run sacrificed %d permanents, want both (%v)", out.got.Count(), out.got)
	}
	if len(out.got) != 1 {
		t.Errorf("two prompts for one seat are one entry, got %d", len(out.got))
	}
}

// TestAFireAndForgetSacrificePromptStillSacrifices guards the ten
// catalog callers that have nothing hanging off the answer: the run
// link is uuid.Nil for them and the answer behaves exactly as it did
// before #1019.
func TestAFireAndForgetSacrificePromptStillSacrifices(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	bear := pushBear(g, me.ID)

	var queued int
	g.WithWriteLock(func() {
		queued = g.PlayerSacrificesForEffect(uuid.Nil, me.ID, nil, "sacrifice a permanent")
	})
	if queued != 1 {
		t.Fatalf("PlayerSacrificesForEffect queued %d prompts, want 1", queued)
	}
	answerSacrifice(t, g, me.ID, bear)

	if !me.Graveyard.Contains(bear) {
		t.Error("a prompted sacrifice with no continuation still puts the permanent in the graveyard")
	}
	if !hasEvent(g, EventSacrifice, bear) {
		t.Error("the announcement still fires (CR 701.17a)")
	}
	if g.promptRuns != nil {
		t.Errorf("a prompt nobody is waiting on left %d runs behind", len(g.promptRuns))
	}
}
