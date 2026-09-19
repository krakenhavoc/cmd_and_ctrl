package game

import (
	"testing"

	"github.com/google/uuid"
)

// sacrifice_run_test.go pins #1019: the prompted sacrifice has a
// continuation, and it runs once, after the LAST asked seat has
// settled.
//
// The two entry points do not sacrifice anything — they queue a
// question per seat and return how many seats were asked — so every
// clause a card wrote on the next line was a payout on a move that had
// not happened. Rise of the Witch-king returned a permanent from the
// graveyard before anybody had chosen a creature.
//
// What a run has to survive, one test each below: several seats
// answering in any order, a leg the CR 903.9 window has PAUSED, a seat
// that leaves mid-prompt, a seat that had nothing to sacrifice, a
// prompt the engine withdraws, and an undo across a half-answered
// fan-out.

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

// answerSacrifice answers the open sacrifice prompt owed by `seat`
// with `card`.
func answerSacrifice(t *testing.T, g *Game, seat, card uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceSacrifice && c.Chooser == seat {
			if err := g.ResolveSacrificeChoice(c.ID, seat, card); err != nil {
				t.Fatalf("ResolveSacrificeChoice: %v", err)
			}
			return
		}
	}
	t.Fatalf("seat %s owes no sacrifice prompt", seat)
}

// openSacrificePrompts counts the sacrifice prompts still queued.
func openSacrificePrompts(g *Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceSacrifice {
			n++
		}
	}
	return n
}

// TestAPromptedSacrificeRunWaitsForEveryAskedSeat is the headline, and
// Rise of the Witch-king's shape: three seats are asked, and the
// continuation runs exactly once — after the third answer, not after
// the third QUESTION.
func TestAPromptedSacrificeRunWaitsForEveryAskedSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	seats := g.Seats[:3]
	source := departureTestSource(g, seats[0].ID, "Rise of the Witch-king")
	bears := map[uuid.UUID]uuid.UUID{}
	for _, p := range seats {
		bears[p.ID] = pushBear(g, p.ID)
	}

	run := eachPlayerSacrificesThen(t, g, source, g.Seats[3].ID)

	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with three prompts open, want 0 — "+
			"queueing a question is not sacrificing anything", run.ran)
	}
	if openSacrificePrompts(g) != 3 {
		t.Fatalf("expected three prompts, got %d", openSacrificePrompts(g))
	}

	// Answered out of ask order on purpose: the run is a join, not a
	// sequence.
	for _, i := range []int{2, 0, 1} {
		seat := seats[i].ID
		answerSacrifice(t, g, seat, bears[seat])
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
	// The ask order is APNAP from the active player, whatever order the
	// answers arrived in: a continuation that walked the landed map
	// would decide differently on two runs of one game.
	for i, e := range run.got {
		if e.Seat != seats[i].ID {
			t.Errorf("entry %d is seat %s, want the ask order (%s)", i, e.Seat, seats[i].Name)
		}
	}
}

// TestAPromptedSacrificeRunWaitsForAPausedCommanderLeg is the CR 903.9
// half, and the reason the answer goes through SacrificeThenForEffect
// rather than the fire-and-forget call: a sacrificed commander is
// STILL ON THE BATTLEFIELD while its owner is asked about the command
// zone, so a run that settled on the line after the answer would pay
// out with the permanent still in play.
func TestAPromptedSacrificeRunWaitsForAPausedCommanderLeg(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Rise of the Witch-king")
	commander := seatCommander(t, g.Battlefield, me)

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
// with no legal permanent is skipped at queue time (CR 701.21a's "if
// you can"), so it is never part of the run and the run does not wait
// for it.
func TestAPromptedSacrificeRunIgnoresASeatWithNothingToSacrifice(t *testing.T) {
	g := newActiveGame(t)
	me, empty := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Fleshbag Marauder")
	bear := pushBear(g, me.ID)

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

// TestAWithdrawnSacrificePromptSettlesItsLegWithNothing is the
// dropDefault half (#1016's rule at a second kind): the engine
// withdraws a prompt whose board has emptied under it, and the run
// hears "this seat sacrificed nothing" rather than waiting forever.
//
// pruneSacrificeChoicesLocked is the real caller — one seat's answer
// can kill another seat's last creature — and it reaches
// dropChoiceLocked, which is the door every prune in the tree goes
// through.
func TestAWithdrawnSacrificePromptSettlesItsLegWithNothing(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Fleshbag Marauder")
	mine := pushBear(g, me.ID)
	theirs := pushBear(g, them.ID)

	run := eachPlayerSacrificesThen(t, g, source, uuid.Nil)
	if openSacrificePrompts(g) != 2 {
		t.Fatalf("expected two prompts, got %d", openSacrificePrompts(g))
	}

	// Their creature leaves by some other route, so the prune has no
	// legal answer left to offer them.
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(theirs); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
		g.pruneSacrificeChoicesLocked()
	})
	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with one prompt still open, want 0", run.ran)
	}

	answerSacrifice(t, g, me.ID, mine)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — a withdrawn leg must settle, "+
			"or the run waits for an answer that is never coming", run.ran)
	}
	if !run.got.Sacrificed(me.ID) {
		t.Error("the seat that answered sacrificed its bear")
	}
	if run.got.Sacrificed(them.ID) {
		t.Errorf("the withdrawn leg reports a sacrifice: %v", run.got)
	}
	if got := run.got.By(them.ID); got != nil {
		t.Errorf("the withdrawn leg reports %v, want nothing", got)
	}
	// The seat is still IN the answer: it was asked, and "asked and
	// sacrificed nothing" is a different fact from "never asked".
	if len(run.got) != 2 {
		t.Errorf("the run reports %d asked seats, want 2", len(run.got))
	}
}

// TestASeatLeavingMidPromptSettlesItsSacrificeLeg is the departure
// sweep's half, gated exactly like every other dropDefault: the card
// belongs to a seat still at the table, so the drop settles the leg
// and the rest of the instruction finishes for everybody else.
func TestASeatLeavingMidPromptSettlesItsSacrificeLeg(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Rise of the Witch-king")
	mine := pushBear(g, caster.ID)
	pushBear(g, leaver.ID)

	run := eachPlayerSacrificesThen(t, g, source, uuid.Nil)
	// Only the two seats with a permanent were asked.
	if openSacrificePrompts(g) != 2 {
		t.Fatalf("expected two prompts, got %d", openSacrificePrompts(g))
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with the caster's prompt still open, want 0", run.ran)
	}

	answerSacrifice(t, g, caster.ID, mine)

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
// clone contract. A run is shared mutable state — the counter and the
// per-seat landed lists — so an undo that rewound the queue and not
// the run would replay the second answer against a counter that had
// already been decremented and pay out a seat early.
func TestAnUndoAcrossAHalfAnsweredSacrificeRunReplaysIdentically(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Fleshbag Marauder")
	mine := pushBear(g, me.ID)
	theirs := pushBear(g, them.ID)

	run := eachPlayerSacrificesThen(t, g, source, uuid.Nil)

	var snapshot *Game
	g.WithWriteLock(func() { snapshot = g.cloneLocked() })

	answerSacrifice(t, g, me.ID, mine)
	if run.ran != 0 {
		t.Fatalf("one answer of two ran the continuation %d times, want 0", run.ran)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
	if openSacrificePrompts(g) != 2 {
		t.Fatalf("the undo restored %d prompts, want 2", openSacrificePrompts(g))
	}
	if run.ran != 0 {
		t.Fatalf("the undo itself ran the continuation %d times", run.ran)
	}

	// Replay the same answer, then the other seat's.
	answerSacrifice(t, g, me.ID, mine)
	if run.ran != 0 {
		t.Fatalf("the replayed answer ran the continuation %d times, want 0 — the counter "+
			"rewound with the queue", run.ran)
	}
	answerSacrifice(t, g, them.ID, theirs)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1", run.ran)
	}
	if run.got.Count() != 2 {
		t.Errorf("the run sacrificed %d permanents, want 2 — a landed list that grew across "+
			"the undo would report three (%v)", run.got.Count(), run.got)
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
