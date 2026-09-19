package game

import (
	"testing"

	"github.com/google/uuid"
)

// discard_run_test.go pins #1027: the prompted discard has a
// continuation, and it runs once, after the LAST asked seat has
// settled and the cards it named have finished moving.
//
// QueueDiscardChoiceForEffect queues a question over the player's own
// hand and returns the prompt's ID. Nothing has left the hand when it
// returns, and a discarded commander's CR 903.9 prompt can hold the
// batch for another action after that — so every clause a card wrote
// on the next line ran too early. Archon of Cruelty ran its last three
// clauses with two prompts still open.
//
// What a run has to survive, one test each below: several seats
// answering in any order, a leg the CR 903.9 window has PAUSED, a
// madness card that went to exile instead of a graveyard, a seat with
// an empty hand, a seat that leaves mid-prompt, and an undo across a
// half-answered fan-out.

// discardRunCall records one run of a continuation.
type discardRunCall struct {
	ran int
	got PromptedDiscards
}

// eachPlayerDiscardsThen starts a fan-out run over every seat and
// records what its continuation is handed.
func eachPlayerDiscardsThen(t *testing.T, g *Game, source, except uuid.UUID, n int) *discardRunCall {
	t.Helper()
	out := &discardRunCall{}
	g.WithWriteLock(func() {
		err := g.EachPlayerDiscardsThenForEffect(except, DiscardPrompt{Source: source, N: n},
			func(_ *Game, discarded PromptedDiscards) error {
				out.ran++
				out.got = discarded
				return nil
			})
		if err != nil {
			t.Fatalf("EachPlayerDiscardsThenForEffect: %v", err)
		}
	})
	return out
}

// answerDiscard answers the open discard prompt owed by `seat` with
// `cards`.
func answerDiscard(t *testing.T, g *Game, seat uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	c := discardPromptFor(g, seat)
	if c == nil {
		t.Fatalf("seat %s owes no discard prompt", seat)
	}
	if err := g.ResolveChooseCards(c.ID, seat, cards); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// openDiscardPrompts counts the discard prompts still queued.
func openDiscardPrompts(g *Game) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceChooseCards && c.Chooser == c.FromPlayer {
			n++
		}
	}
	return n
}

// oneCardHand empties p's hand and deals it a single card back, so a
// discard of one has exactly one legal answer.
func oneCardHand(t *testing.T, g *Game, p *Player, name string) uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	return handCardForDiscard(t, g, p, name)
}

// TestAPromptedDiscardRunWaitsForEveryAskedSeat is the headline, and
// Syphon Mind's shape: three seats are asked, and the continuation
// runs exactly once — after the third ANSWER, not after the third
// question.
func TestAPromptedDiscardRunWaitsForEveryAskedSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	seats := g.Seats[:3]
	source := departureTestSource(g, seats[0].ID, "Syphon Mind")
	pitched := map[uuid.UUID]uuid.UUID{}
	for _, p := range seats {
		pitched[p.ID] = oneCardHand(t, g, p, "Pitch")
	}

	run := eachPlayerDiscardsThen(t, g, source, g.Seats[3].ID, 1)

	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with three prompts open, want 0 — "+
			"queueing a question is not discarding anything", run.ran)
	}
	if openDiscardPrompts(g) != 3 {
		t.Fatalf("expected three prompts, got %d", openDiscardPrompts(g))
	}

	// Answered out of ask order on purpose: the run is a join, not a
	// sequence.
	for _, i := range []int{2, 0, 1} {
		seat := seats[i].ID
		answerDiscard(t, g, seat, pitched[seat])
	}

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1", run.ran)
	}
	if run.got.Count() != 3 {
		t.Errorf("the run discarded %d cards, want 3 (%v)", run.got.Count(), run.got)
	}
	for _, p := range seats {
		if !run.got.Discarded(p.ID) {
			t.Errorf("seat %s answered and is reported as having discarded nothing", p.Name)
		}
		if got := run.got.By(p.ID); len(got) != 1 || got[0] != pitched[p.ID] {
			t.Errorf("seat %s discarded %v, want the card it named", p.Name, got)
		}
		if !p.Graveyard.Contains(pitched[p.ID]) {
			t.Errorf("seat %s's card is not in its graveyard", p.Name)
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

// TestAPromptedDiscardRunWaitsForAPausedCommanderLeg is the CR 903.9
// half, and the reason the leg settles from inside the discard batch
// rather than on the line after the answer: a discarded commander is
// STILL IN THE HAND while its owner is asked about the command zone.
func TestAPromptedDiscardRunWaitsForAPausedCommanderLeg(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	commander := seatCommander(t, me.Hand, me)

	out := &discardRunCall{}
	g.WithWriteLock(func() {
		err := g.PlayerDiscardsThenForEffect(DiscardPrompt{Player: me.ID, Source: source, N: 1},
			func(_ *Game, discarded PromptedDiscards) error {
				out.ran++
				out.got = discarded
				return nil
			})
		if err != nil {
			t.Fatalf("PlayerDiscardsThenForEffect: %v", err)
		}
	})
	answerDiscard(t, g, me.ID, commander)

	if out.ran != 0 {
		t.Fatalf("the continuation ran %d times with the CR 903.9 prompt open, want 0", out.ran)
	}
	if !me.Hand.Contains(commander) {
		t.Error("a paused leg has moved nothing")
	}

	prompt := expectCommanderPrompt(t, g, me)
	if err := g.ResolveOptionalReplacement(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}

	if out.ran != 1 {
		t.Fatalf("the continuation ran %d times after the CR 903.9 answer, want 1", out.ran)
	}
	if !out.got.Discarded(me.ID) {
		t.Errorf("the commander took the command zone and the run reports nothing discarded: "+
			"CR 701.8a's discard is the move OUT of the hand, and only where the card went "+
			"was replaced (%v)", out.got)
	}
	if !me.Command.Contains(commander) {
		t.Error("the commander is in the command zone")
	}
}

// TestAMadnessExileCountsAsDiscardedThisWay: CR 702.35a says the
// player "discards it, BUT exiles it instead of putting it into their
// graveyard". The discard happened; only the destination was
// rewritten — so Syphon Mind draws for it.
func TestAMadnessExileCountsAsDiscardedThisWay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	withMadnessCard(t, "{R}")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	out := &discardRunCall{}
	g.WithWriteLock(func() {
		err := g.PlayerDiscardsThenForEffect(DiscardPrompt{Player: me.ID, Source: source, N: 1},
			func(_ *Game, discarded PromptedDiscards) error {
				out.ran++
				out.got = discarded
				return nil
			})
		if err != nil {
			t.Fatalf("PlayerDiscardsThenForEffect: %v", err)
		}
	})
	answerDiscard(t, g, me.ID, card)

	if out.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", out.ran)
	}
	if inGraveyard(me, card) {
		t.Fatal("the madness card went to the graveyard; CR 702.35a exiles it instead")
	}
	if exiledCardByIDLocked(g, card) == nil {
		t.Fatal("the madness card is not in exile")
	}
	if !out.got.Discarded(me.ID) {
		t.Errorf("a madness card exiled instead of binned is still DISCARDED (CR 702.35a) — "+
			"the run reports %v", out.got)
	}
	if got := out.got.By(me.ID); len(got) != 1 || got[0] != card {
		t.Errorf("the run reports %v, want the exiled card", got)
	}
}

// TestACancelledDiscardIsNotDiscardedThisWay is the other side of the
// same rule: a replacement that leaves the card in the hand means
// nothing was discarded, however thoroughly the prompt was answered.
func TestACancelledDiscardIsNotDiscardedThisWay(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	card := oneCardHand(t, g, me, "Pitch")
	// "Instead, it stays where it is" — the CR 614 cancel, spelled as
	// a move back to the zone it is already in.
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(discardReplacement("keep it instead?", ZoneHand, false, DiscardCauseEffect))
	})

	out := &discardRunCall{}
	g.WithWriteLock(func() {
		err := g.PlayerDiscardsThenForEffect(DiscardPrompt{Player: me.ID, Source: source, N: 1},
			func(_ *Game, discarded PromptedDiscards) error {
				out.ran++
				out.got = discarded
				return nil
			})
		if err != nil {
			t.Fatalf("PlayerDiscardsThenForEffect: %v", err)
		}
	})
	answerDiscard(t, g, me.ID, card)

	if out.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — a cancelled leg still settles", out.ran)
	}
	if !me.Hand.Contains(card) {
		t.Fatal("the replacement kept the card in the hand")
	}
	if out.got.Discarded(me.ID) {
		t.Errorf("a card that never left the hand is reported as discarded: %v", out.got)
	}
	if len(out.got) != 1 {
		t.Errorf("the seat was ASKED, so it has an entry: %v", out.got)
	}
}

// TestAPromptedDiscardRunIgnoresASeatWithAnEmptyHand: CR 701.8a
// discards as many as you can, which for an empty hand is nothing. No
// prompt goes up, so the seat is not part of the run and the run does
// not wait for it.
func TestAPromptedDiscardRunIgnoresASeatWithAnEmptyHand(t *testing.T) {
	g := newActiveGame(t)
	me, empty := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Syphon Mind")
	card := oneCardHand(t, g, me, "Pitch")
	g.WithWriteLock(func() { empty.Hand.Cards = nil })

	run := eachPlayerDiscardsThen(t, g, source, uuid.Nil, 1)

	if openDiscardPrompts(g) != 1 {
		t.Fatalf("expected one prompt — the other seat is empty-handed — got %d",
			openDiscardPrompts(g))
	}
	answerDiscard(t, g, me.ID, card)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", run.ran)
	}
	if len(run.got) != 1 || run.got[0].Seat != me.ID {
		t.Errorf("the run reports %v, want one entry for the seat that was ASKED", run.got)
	}
	if run.got.Discarded(empty.ID) {
		t.Error("an empty-handed seat is reported as having discarded something")
	}
	if run.got.Count() != 1 {
		t.Errorf("the run discarded %d cards, want 1", run.got.Count())
	}
}

// TestAPromptedDiscardRunWithNobodyToAskStillRunsTheRestOfTheCard is
// the #544 rule at queue time: a continuation is the rest of a card
// that is paused mid-resolution, so an instruction that could not ask
// anybody still finishes. Syphon Mind against a table of empty hands
// draws nothing and does not stop halfway.
func TestAPromptedDiscardRunWithNobodyToAskStillRunsTheRestOfTheCard(t *testing.T) {
	g := newActiveGame(t)
	source := departureTestSource(g, g.Seats[0].ID, "Syphon Mind")
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
	})

	run := eachPlayerDiscardsThen(t, g, source, uuid.Nil, 1)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times with nobody to ask, want 1", run.ran)
	}
	if run.got.Count() != 0 {
		t.Errorf("nothing was discarded, the run reports %v", run.got)
	}
}

// TestTheLegsOwnThenStillRunsForAnEmptyHand guards the contract
// DiscardPrompt.Then has had since #797, which the run must not
// change: "discard your hand, THEN draw three" draws three from an
// empty hand.
func TestTheLegsOwnThenStillRunsForAnEmptyHand(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	legs, run := 0, 0
	g.WithWriteLock(func() {
		err := g.PlayerDiscardsThenForEffect(
			DiscardPrompt{
				Player: me.ID,
				N:      1,
				Then:   func(*Game, uuid.UUID, []uuid.UUID) error { legs++; return nil },
			},
			func(_ *Game, _ PromptedDiscards) error { run++; return nil })
		if err != nil {
			t.Fatalf("PlayerDiscardsThenForEffect: %v", err)
		}
	})

	if legs != 1 {
		t.Errorf("the prompt's own Then ran %d times, want 1", legs)
	}
	if run != 1 {
		t.Errorf("the run's continuation ran %d times, want 1", run)
	}
}

// TestASeatLeavingMidPromptSettlesItsDiscardLeg is the dropDefault
// half at the third kind (#1016's rule, #1027's currency): a seat that
// leaves with a discard prompt open settles its leg with nothing, and
// the rest of the instruction finishes for everybody else.
//
// The prompt is a PendingChoiceChooseCards over the departing seat's
// OWN hand, so it is never reassigned (gate 3), and the drop is gated
// on the asking card surviving with its controller (gate 2) exactly as
// every other dropDefault is.
func TestASeatLeavingMidPromptSettlesItsDiscardLeg(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Syphon Mind")
	mine := oneCardHand(t, g, caster, "Pitch")
	oneCardHand(t, g, leaver, "Pitch")
	g.WithWriteLock(func() {
		g.Seats[2].Hand.Cards = nil
		g.Seats[3].Hand.Cards = nil
	})

	run := eachPlayerDiscardsThen(t, g, source, uuid.Nil, 1)
	if openDiscardPrompts(g) != 2 {
		t.Fatalf("expected two prompts, got %d", openDiscardPrompts(g))
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with the caster's prompt still open, want 0", run.ran)
	}

	answerDiscard(t, g, caster.ID, mine)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — a seat that leaves mid-prompt must "+
			"not strand the run", run.ran)
	}
	if !run.got.Discarded(caster.ID) {
		t.Error("the surviving seat discarded its card")
	}
	if run.got.Discarded(leaver.ID) {
		t.Errorf("the departed seat is reported as having discarded something: %v", run.got)
	}
	if run.got.Count() != 1 {
		t.Errorf("the run discarded %d cards, want 1", run.got.Count())
	}
}

// TestAnUndoAcrossAHalfAnsweredDiscardRunReplaysIdentically is the
// clone contract. A run is shared mutable state — the counter and the
// per-seat landed lists — so an undo that rewound the queue and not
// the run would replay the second answer against a counter that had
// already been decremented and pay out a seat early.
func TestAnUndoAcrossAHalfAnsweredDiscardRunReplaysIdentically(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Syphon Mind")
	mine := oneCardHand(t, g, me, "Mine")
	theirs := oneCardHand(t, g, them, "Theirs")

	run := eachPlayerDiscardsThen(t, g, source, uuid.Nil, 1)

	var snapshot *Game
	g.WithWriteLock(func() { snapshot = g.cloneLocked() })

	answerDiscard(t, g, me.ID, mine)
	if run.ran != 0 {
		t.Fatalf("one answer of two ran the continuation %d times, want 0", run.ran)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snapshot) })
	if openDiscardPrompts(g) != 2 {
		t.Fatalf("the undo restored %d prompts, want 2", openDiscardPrompts(g))
	}
	if run.ran != 0 {
		t.Fatalf("the undo itself ran the continuation %d times", run.ran)
	}

	// Replay the same answer, then the other seat's.
	answerDiscard(t, g, me.ID, mine)
	if run.ran != 0 {
		t.Fatalf("the replayed answer ran the continuation %d times, want 0 — the counter "+
			"rewound with the queue", run.ran)
	}
	answerDiscard(t, g, them.ID, theirs)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1", run.ran)
	}
	if run.got.Count() != 2 {
		t.Errorf("the run discarded %d cards, want 2 — a landed list that grew across the "+
			"undo would report three (%v)", run.got.Count(), run.got)
	}
	if got := run.got.By(me.ID); len(got) != 1 {
		t.Errorf("the replayed seat discarded %v, want exactly one card", got)
	}
}

// TestPlayerDiscardsThenCountsTheWholeInstruction: "discard two cards"
// is ONE prompt over the hand, and the run's answer carries both
// cards.
func TestPlayerDiscardsThenCountsTheWholeInstruction(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	first := handCardForDiscard(t, g, me, "First")
	second := handCardForDiscard(t, g, me, "Second")

	out := &discardRunCall{}
	g.WithWriteLock(func() {
		err := g.PlayerDiscardsThenForEffect(DiscardPrompt{Player: me.ID, Source: source, N: 2},
			func(_ *Game, discarded PromptedDiscards) error {
				out.ran++
				out.got = discarded
				return nil
			})
		if err != nil {
			t.Fatalf("PlayerDiscardsThenForEffect: %v", err)
		}
	})
	if openDiscardPrompts(g) != 1 {
		t.Fatalf("a discard of two is one question, got %d prompts", openDiscardPrompts(g))
	}
	answerDiscard(t, g, me.ID, first, second)

	if out.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", out.ran)
	}
	if out.got.Count() != 2 {
		t.Errorf("the run discarded %d cards, want both (%v)", out.got.Count(), out.got)
	}
	if len(out.got) != 1 {
		t.Errorf("one seat is one entry, got %d", len(out.got))
	}
}

// TestAFireAndForgetDiscardPromptStillDiscards guards the twenty-odd
// catalog callers that have nothing hanging off the answer: the run
// link is uuid.Nil for them and the answer behaves exactly as it did
// before #1027.
func TestAFireAndForgetDiscardPromptStillDiscards(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	card := oneCardHand(t, g, me, "Pitch")

	if id := queueEffectDiscard(g, DiscardPrompt{Player: me.ID, N: 1}); id == uuid.Nil {
		t.Fatal("QueueDiscardChoiceForEffect queued nothing against a one-card hand")
	}
	answerDiscard(t, g, me.ID, card)

	if !me.Graveyard.Contains(card) {
		t.Error("a discard prompt with no continuation still puts the card in the graveyard")
	}
	if g.promptRuns != nil {
		t.Errorf("a prompt nobody is waiting on left %d runs behind", len(g.promptRuns))
	}
}

// TestADiscardRunFinishingClearsTheRegistry: the run entry lives
// exactly as long as the instruction. A registry that leaked would put
// a continuation frame on the census for every Mind Rot, and would
// hand an undo snapshot runs nobody is waiting on.
func TestADiscardRunFinishingClearsTheRegistry(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	card := oneCardHand(t, g, me, "Pitch")

	g.WithWriteLock(func() {
		err := g.PlayerDiscardsThenForEffect(DiscardPrompt{Player: me.ID, Source: source, N: 1},
			func(*Game, PromptedDiscards) error { return nil })
		if err != nil {
			t.Fatalf("PlayerDiscardsThenForEffect: %v", err)
		}
	})
	if len(g.promptRuns) != 1 {
		t.Fatalf("an open instruction holds %d runs, want 1", len(g.promptRuns))
	}
	answerDiscard(t, g, me.ID, card)
	if g.promptRuns != nil {
		t.Errorf("a finished instruction left %d runs behind", len(g.promptRuns))
	}
}
