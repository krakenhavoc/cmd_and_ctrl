package game

import (
	"testing"

	"github.com/google/uuid"
)

// discard_commander_test.go pins #853: a DISCARD is an exit like any
// other, so a discarded commander is offered the command zone.
//
// Before #853 every discard path moved the card hand → graveyard with
// a raw MoveCard — the last exit in the engine that did not go through
// routeCardToZoneLocked — and a Mind Rot on your commander put it in
// the bin and asked nobody. ADR 0013 §5f recorded the gap.
//
// Since ADR 0115 a discarded commander is discarded like any other
// card: it goes to its owner's graveyard, EventDiscardCard names the
// graveyard, the discard's "then" runs inline, and the CR 903.9a
// state-based action asks the owner afterwards (commander_return).
// Every test here asserts both halves: the card lands in the graveyard
// first, AND it ends where the owner's answer says.

// discardWatcher records what the discard actually emitted. A discard
// has always been exactly one event — EventDiscardCard, never an
// EventZoneMove as well — and Syr Konrad and Bloodchief Ascension both
// watch the whole family, so a second event would double-count.
type discardWatcher struct {
	discards  []Event
	zoneMoves []Event
}

func (w *discardWatcher) OnEvent(_ *Game, ev Event) {
	switch ev.Kind {
	case EventDiscardCard:
		w.discards = append(w.discards, ev)
	case EventZoneMove:
		w.zoneMoves = append(w.zoneMoves, ev)
	}
}

// watchDiscards registers a recorder on g and returns it.
func watchDiscards(g *Game) *discardWatcher {
	w := &discardWatcher{}
	g.RegisterListener(w)
	return w
}

// emptyHandWithCommanders clears p's hand and deals it n commanders,
// returning their IDs in hand order.
func emptyHandWithCommanders(t *testing.T, g *Game, p *Player, n int) []uuid.UUID {
	t.Helper()
	g.WithWriteLock(func() { p.Hand.Cards = nil })
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, seatCommander(t, p.Hand, p))
	}
	return ids
}

// commanderPromptCard reports which card the queued CR 903.9 prompt is
// asking about, read off the resume frame the prompt is holding.
func commanderPromptCard(t *testing.T, c *PendingChoice) uuid.UUID {
	t.Helper()
	if c == nil || c.replacementResume == nil || c.replacementResume.ev == nil {
		t.Fatal("the CR 903.9 prompt carries no replacement resume frame")
	}
	return c.replacementResume.ev.CardID
}

// assertOneDiscardEvent checks the single event a one-card discard
// owes: the discarding player, the card, out of the hand, and into the
// zone it landed in. A commander's later CR 903.9a move from the
// graveyard to the command zone is its own EventZoneMove, out of the
// graveyard, and is not counted against the discard.
func assertOneDiscardEvent(t *testing.T, w *discardWatcher, actor, cardID uuid.UUID, landed ZoneKind) {
	t.Helper()
	if len(w.discards) != 1 {
		t.Fatalf("EventDiscardCard x %d, want 1", len(w.discards))
	}
	ev := w.discards[0]
	if ev.Actor != actor || ev.CardID != cardID {
		t.Errorf("EventDiscardCard actor/card = %s/%s, want %s/%s", ev.Actor, ev.CardID, actor, cardID)
	}
	if ev.OldZone != ZoneHand || ev.NewZone != landed {
		t.Errorf("EventDiscardCard zones %s → %s, want hand → %s", ev.OldZone, ev.NewZone, landed)
	}
	for _, zm := range w.zoneMoves {
		if zm.CardID == cardID && zm.OldZone == ZoneHand {
			t.Errorf("the discard also emitted EventZoneMove — Syr Konrad counts it twice")
		}
	}
}

// --- (b) the effect discard: Mind Rot on a commander -----------------

// TestEffectDiscardOfACommanderOffersTheCommandZone is the issue, on
// the path most discards take: QueueDiscardChoiceForEffect's
// continuation (#797). The commander is discarded into the graveyard,
// the prompt's own "then" runs at once, and CR 903.9a then asks. Both
// answers.
func TestEffectDiscardOfACommanderOffersTheCommandZone(t *testing.T) {
	for _, toCommandZone := range []bool{true, false} {
		name := "declined"
		if toCommandZone {
			name = "command zone"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			p := g.Seats[1]
			cmdID := emptyHandWithCommanders(t, g, p, 1)[0]
			w := watchDiscards(g)

			thenRuns := 0
			if id := queueEffectDiscard(g, DiscardPrompt{
				Player: p.ID,
				N:      1,
				Then:   func(*Game, uuid.UUID, []uuid.UUID) error { thenRuns++; return nil },
			}); id == uuid.Nil {
				t.Fatal("a one-card discard against a one-card hand must queue a prompt")
			}
			c := discardPromptFor(g, p.ID)
			if c == nil {
				t.Fatal("no discard prompt queued")
			}
			if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{cmdID}); err != nil {
				t.Fatalf("ResolveChooseCards: %v", err)
			}

			// Nothing waits on the owner: the card is in the graveyard,
			// the discard was emitted, and "then" has run.
			assertOnlyIn(t, cmdID, p.Graveyard, p.Hand, p.Command)
			assertOneDiscardEvent(t, w, p.ID, cmdID, ZoneGraveyard)
			if thenRuns != 1 {
				t.Errorf(`"then" ran %d times, want 1 — the discard landed at once`, thenRuns)
			}

			answerCommanderReturn(t, g, p, cmdID, toCommandZone)
			if len(g.PendingChoices) != 0 {
				t.Fatalf("%d prompts survived the answer", len(g.PendingChoices))
			}
			want, other := p.Command, p.Graveyard
			if !toCommandZone {
				want, other = p.Graveyard, p.Command
			}
			assertOnlyIn(t, cmdID, want, other, p.Hand, g.Exile)
			assertOneDiscardEvent(t, w, p.ID, cmdID, ZoneGraveyard)
			if thenRuns != 1 {
				t.Errorf(`"then" ran %d times, want 1`, thenRuns)
			}
		})
	}
}

// --- (f) two commanders at once (partners) ---------------------------

// TestTwoDiscardedCommandersAreAskedOneEachAndRunThenOnce — a partner
// pair pitched to one Mind Rot is discarded as one batch, the prompt's
// own "then" runs once, and the CR 903.9a check then asks one question
// per commander, both open at once (ADR 0115 decision 3).
func TestTwoDiscardedCommandersAreAskedOneEachAndRunThenOnce(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	ids := emptyHandWithCommanders(t, g, p, 2)
	first, second := ids[0], ids[1]
	w := watchDiscards(g)

	thenRuns := 0
	queueEffectDiscard(g, DiscardPrompt{
		Player: p.ID,
		N:      2,
		Then:   func(*Game, uuid.UUID, []uuid.UUID) error { thenRuns++; return nil },
	})
	c := discardPromptFor(g, p.ID)
	if c == nil {
		t.Fatal("no discard prompt queued")
	}
	if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if thenRuns != 1 {
		t.Errorf(`"then" ran %d times, want exactly 1 for the batch`, thenRuns)
	}
	assertOnlyIn(t, first, p.Graveyard, p.Hand)
	assertOnlyIn(t, second, p.Graveyard, p.Hand)

	runChecks(g)
	prompts := commanderReturnPrompts(g)
	if len(prompts) != 2 || len(g.PendingChoices) != 2 {
		t.Fatalf("commander_return prompts = %d of %d pending, want 2", len(prompts), len(g.PendingChoices))
	}
	answers := map[uuid.UUID]bool{first: true, second: false}
	for _, pr := range prompts {
		apply, ok := answers[pr.Source]
		if !ok || pr.Chooser != p.ID {
			t.Fatalf("prompt to %s about %s, want the owner about one of the two", pr.Chooser, pr.Source)
		}
		if err := g.ResolveCommanderReturn(pr.ID, p.ID, apply); err != nil {
			t.Fatalf("ResolveCommanderReturn: %v", err)
		}
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the batch", len(g.PendingChoices))
	}
	assertOnlyIn(t, first, p.Command, p.Graveyard, p.Hand)
	assertOnlyIn(t, second, p.Graveyard, p.Command, p.Hand)

	// (h) one discard event per card, both into the graveyard — the
	// "whenever you discard a card" family fires twice.
	if len(w.discards) != 2 {
		t.Fatalf("EventDiscardCard x %d, want 2 (one per card)", len(w.discards))
	}
	for _, ev := range w.discards {
		if ev.Actor != p.ID || ev.OldZone != ZoneHand || ev.NewZone != ZoneGraveyard {
			t.Errorf("EventDiscardCard actor/zones = %s/%s → %s, want %s/hand → graveyard", ev.Actor, ev.OldZone, ev.NewZone, p.ID)
		}
	}
}

// --- (d) the random discard ------------------------------------------

// TestRandomDiscardOfACommanderOffersTheCommandZone — CR 701.9b picks
// the card, and CR 903.9a asks about it once it is in the graveyard. A
// one-card hand makes the random pick the commander.
func TestRandomDiscardOfACommanderOffersTheCommandZone(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	cmdID := emptyHandWithCommanders(t, g, p, 1)[0]
	w := watchDiscards(g)

	var err error
	g.WithWriteLock(func() { err = g.DiscardRandomForEffect(p.ID, 1) })
	if err != nil {
		t.Fatalf("DiscardRandomForEffect: %v", err)
	}
	assertOnlyIn(t, cmdID, p.Graveyard, p.Hand)
	answerCommanderReturn(t, g, p, cmdID, true)
	assertOnlyIn(t, cmdID, p.Command, p.Graveyard, p.Hand)
	assertOneDiscardEvent(t, w, p.ID, cmdID, ZoneGraveyard)
}

// --- (c) the cleanup discard -----------------------------------------

// TestCleanupDiscardOfACommanderOffersTheCommandZone — the CR 514.1
// hand-size discard is a turn-based action, not an effect. The
// commander is discarded with the rest and the debt is paid at once.
// Then the cleanup step's CR 514.3a check asks the owner: a "yes" is a
// state-based action performed, so the active player gets priority in
// the cleanup step; a "no" performs nothing, and the turn ends.
func TestCleanupDiscardOfACommanderOffersTheCommandZone(t *testing.T) {
	for _, toCommandZone := range []bool{true, false} {
		name := "declined"
		if toCommandZone {
			name = "command zone"
		}
		t.Run(name, func(t *testing.T) {
			g := newActiveGame(t)
			active := g.Seats[0]
			cmdID := seatCommander(t, active.Hand, active)
			for active.Hand.Size() < 9 {
				if err := g.DrawCard(active.ID); err != nil {
					t.Fatalf("DrawCard: %v", err)
				}
			}
			spare := active.Hand.Cards[0].InstanceID
			if spare == cmdID {
				t.Fatal("the spare pick must not be the commander")
			}
			advanceTo(t, g, StepEnd)
			if _, err := g.AdvanceStep(); err != nil {
				t.Fatalf("AdvanceStep past End: %v", err)
			}
			if g.Turn.Step != StepCleanup {
				t.Fatalf("expected to pause at Cleanup, got %q", g.Turn.Step)
			}
			if g.DiscardPending[active.ID] != 2 {
				t.Fatalf("hand-size debt = %d, want 2", g.DiscardPending[active.ID])
			}
			w := watchDiscards(g)

			if err := g.DiscardSelection(active.ID, []uuid.UUID{cmdID, spare}); err != nil {
				t.Fatalf("DiscardSelection: %v", err)
			}
			assertOnlyIn(t, cmdID, active.Graveyard, active.Hand)
			assertOnlyIn(t, spare, active.Graveyard, active.Hand)
			if len(w.discards) != 2 {
				t.Errorf("EventDiscardCard x %d, want 2", len(w.discards))
			}
			if len(g.DiscardPending) != 0 {
				t.Errorf("DiscardPending survived the cleanup discard: %+v", g.DiscardPending)
			}

			prompts := commanderReturnPrompts(g)
			if len(prompts) != 1 || prompts[0].Source != cmdID || prompts[0].Chooser != active.ID {
				t.Fatalf("cleanup asked %+v, want one CR 903.9a question about the commander", prompts)
			}
			if g.Turn.Step != StepCleanup || g.Turn.PriorityHolder != NoPriority {
				t.Fatalf("step %v, priority %d while the question is open; want cleanup, nobody", g.Turn.Step, g.Turn.PriorityHolder)
			}
			if err := g.ResolveCommanderReturn(prompts[0].ID, active.ID, toCommandZone); err != nil {
				t.Fatalf("ResolveCommanderReturn: %v", err)
			}

			want, other := active.Command, active.Graveyard
			if !toCommandZone {
				want, other = active.Graveyard, active.Command
			}
			assertOnlyIn(t, cmdID, want, other, active.Hand)
			if toCommandZone {
				if g.Turn.Step != StepCleanup || g.Turn.PriorityHolder != 0 {
					t.Errorf("after yes: step %v, priority %d; want cleanup with priority to seat 0 (CR 514.3a)", g.Turn.Step, g.Turn.PriorityHolder)
				}
				return
			}
			if g.Turn.ActiveSeat == 0 {
				t.Error("after no: the cursor stayed on seat 0; the cleanup step should have ended the turn")
			}
		})
	}
}

// --- (e) the cost discard ---------------------------------------------

// TestCostDiscardOfACommanderIsPaidThenOffered — CR 601.2h pays a
// spell's costs as one indivisible step. Until #1397 a discarded
// commander's owner was never asked at all; #1397 asked BEFORE the
// cast was paid for. Since ADR 0115 the discard is paid like any other:
// the commander goes to the graveyard with the spell already cast, and
// CR 903.9a asks afterwards. The ask-first gate is left for costs that
// put a card into a hand or a library (CR 903.9b,
// cost_commander_choice_test.go).
func TestCostDiscardOfACommanderIsPaidThenOffered(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-thrill"
	withCatalogAdditionalCost(t, func(id string) *AdditionalCost {
		if id == oracle {
			return &AdditionalCost{DiscardCards: 1, Label: "Discard a card"}
		}
		return nil
	})
	spell, _ := thrillInHand(t, g, me, oracle, 0)
	cmdID := seatCommander(t, me.Hand, me)
	w := watchDiscards(g)

	if err := g.CastSpell(me.ID, spell, CastSpellParams{DiscardIDs: []uuid.UUID{cmdID}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !g.Stack.Contains(spell) {
		t.Fatal("the spell was not cast")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceOptionalReplacement {
			t.Fatal("a discard cost asked the CR 903.9b replacement before paying")
		}
	}
	assertOnlyIn(t, cmdID, me.Graveyard, me.Hand, me.Command)
	assertOneDiscardEvent(t, w, me.ID, cmdID, ZoneGraveyard)

	answerCommanderReturn(t, g, me, cmdID, true)
	assertOnlyIn(t, cmdID, me.Command, me.Graveyard, me.Hand)
	if !g.Stack.Contains(spell) {
		t.Error("answering the commander's question took the spell off the stack")
	}
}

// --- (g) undo into the open question ---------------------------------

// TestUndoIntoAnOpenCommanderReturnReplaysTheSameWay — ADR 0115 §8:
// undo inside an open commander_return prompt rewinds to the board the
// owner saw, with the commander and the other card in the graveyard
// and the question open, and answering again comes out the same.
//
// This replaced #853's undo-across-the-discard-pause test: a discard no
// longer pauses on the commander, so there is no held continuation left
// to rewind across.
func TestUndoIntoAnOpenCommanderReturnReplaysTheSameWay(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	cmdID := emptyHandWithCommanders(t, g, p, 1)[0]
	spare := NewCard("Spare", p.ID)
	spare.TypeLine = "Sorcery"
	g.WithWriteLock(func() { p.Hand.PushTop(spare) })

	thenRuns := 0
	queueEffectDiscard(g, DiscardPrompt{
		Player: p.ID,
		N:      2,
		Then:   func(*Game, uuid.UUID, []uuid.UUID) error { thenRuns++; return nil },
	})
	c := discardPromptFor(g, p.ID)
	if c == nil {
		t.Fatal("no discard prompt queued")
	}
	if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{cmdID, spare.InstanceID}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if thenRuns != 1 {
		t.Fatalf(`"then" ran %d times, want 1`, thenRuns)
	}
	runChecks(g)
	promptOpen := g.Clone()

	answerCommanderReturn(t, g, g.Seats[1], cmdID, true)
	assertOnlyIn(t, cmdID, g.Seats[1].Command, g.Seats[1].Graveyard, g.Seats[1].Hand)
	assertOnlyIn(t, spare.InstanceID, g.Seats[1].Graveyard, g.Seats[1].Hand)

	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Seats[1].Graveyard.Contains(cmdID) || !g.Seats[1].Graveyard.Contains(spare.InstanceID) {
		t.Fatal("the rewind into the open question did not put both cards back in the graveyard")
	}
	answerCommanderReturn(t, g, g.Seats[1], cmdID, true)
	if thenRuns != 1 {
		t.Errorf(`replay: "then" ran %d times in all, want 1 — it ran before the question`, thenRuns)
	}
	assertOnlyIn(t, cmdID, g.Seats[1].Command, g.Seats[1].Graveyard, g.Seats[1].Hand)
	assertOnlyIn(t, spare.InstanceID, g.Seats[1].Graveyard, g.Seats[1].Hand)
}

// answerOnlyCommanderPrompt answers the single queued CR 903.9b
// replacement prompt (a bounce or a tuck since ADR 0115).
func answerOnlyCommanderPrompt(t *testing.T, g *Game, chooser uuid.UUID, apply bool) {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want exactly one CR 903.9 prompt", len(g.PendingChoices))
	}
	c := g.PendingChoices[0]
	if c.Kind != PendingChoiceOptionalReplacement {
		t.Fatalf("prompt kind = %q, want %q", c.Kind, PendingChoiceOptionalReplacement)
	}
	if err := g.ResolveOptionalReplacement(c.ID, chooser, apply); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
}

// --- (h) an ordinary discard is untouched -----------------------------

// TestOrdinaryDiscardStillEmitsOneEventPerCardAndNeverPauses — the
// regression guard on the 99% case. A non-commander discard goes
// through the same window and comes out exactly as it did before:
// straight to the graveyard, one EventDiscardCard per card, no
// EventZoneMove, no prompt, and "then" inline.
func TestOrdinaryDiscardStillEmitsOneEventPerCardAndNeverPauses(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	hand := p.Hand.Size()
	if hand < 2 {
		t.Fatalf("opening hand is %d; this test needs two cards", hand)
	}
	picks := []uuid.UUID{p.Hand.Cards[0].InstanceID, p.Hand.Cards[1].InstanceID}
	w := watchDiscards(g)

	thenRuns := 0
	queueEffectDiscard(g, DiscardPrompt{
		Player: p.ID,
		N:      2,
		Then:   func(*Game, uuid.UUID, []uuid.UUID) error { thenRuns++; return nil },
	})
	c := discardPromptFor(g, p.ID)
	if c == nil {
		t.Fatal("no discard prompt queued")
	}
	if err := g.ResolveChooseCards(c.ID, p.ID, picks); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	if len(g.PendingChoices) != 0 {
		t.Fatalf("an ordinary discard queued %d prompt(s)", len(g.PendingChoices))
	}
	if thenRuns != 1 {
		t.Errorf(`"then" ran %d times, want 1 — inline, on the unpaused path`, thenRuns)
	}
	for _, id := range picks {
		assertOnlyIn(t, id, p.Graveyard, p.Hand, p.Command)
	}
	if len(w.discards) != 2 {
		t.Fatalf("EventDiscardCard x %d, want 2", len(w.discards))
	}
	for _, ev := range w.discards {
		if ev.OldZone != ZoneHand || ev.NewZone != ZoneGraveyard {
			t.Errorf("EventDiscardCard zones %s → %s, want hand → graveyard", ev.OldZone, ev.NewZone)
		}
	}
	for _, zm := range w.zoneMoves {
		for _, id := range picks {
			if zm.CardID == id {
				t.Error("a discard emitted EventZoneMove as well — Syr Konrad counts it twice")
			}
		}
	}
}
