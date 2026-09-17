package game

import (
	"testing"

	"github.com/google/uuid"
)

// discard_commander_test.go pins #853: a DISCARD is an exit like any
// other, so a discarded commander gets CR 903.9's "put it into the
// command zone instead" offer.
//
// Before the fix every discard path moved the card hand → graveyard
// with a raw MoveCard — the last exit in the engine that did not go
// through routeCardToZoneLocked, and therefore the last one that never
// opened the CR 614 window. Destroy, exile, mill, counter and bounce
// have all offered it since #539/#851; a Mind Rot on your commander
// put it in the bin and asked nobody. ADR 0013 §5f recorded the gap
// and did not fix it.
//
// Every test here asserts BOTH halves, the way commander_zone_routes_test.go
// does: the prompt is offered, AND the card lands where the answer
// says it should. It also asserts the halves that are new to a
// discard — that EventDiscardCard fires either way (CR 701.8a: a card
// put into the command zone instead was still discarded), that a
// multi-card batch asks once per commander and runs its "then" once at
// the end, and that a COST never asks at all (CR 601.2h).

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
// owes: the discarding player, the card, out of the hand, and into
// whatever zone the CR 903.9 answer settled on.
func assertOneDiscardEvent(t *testing.T, w *discardWatcher, actor, cardID uuid.UUID, landed ZoneKind) {
	t.Helper()
	if len(w.discards) != 1 {
		t.Fatalf("EventDiscardCard x %d, want 1 — a commander put into the command zone was still discarded (CR 701.8a)", len(w.discards))
	}
	ev := w.discards[0]
	if ev.Actor != actor || ev.CardID != cardID {
		t.Errorf("EventDiscardCard actor/card = %s/%s, want %s/%s", ev.Actor, ev.CardID, actor, cardID)
	}
	if ev.OldZone != ZoneHand || ev.NewZone != landed {
		t.Errorf("EventDiscardCard zones %s → %s, want hand → %s", ev.OldZone, ev.NewZone, landed)
	}
	for _, zm := range w.zoneMoves {
		if zm.CardID == cardID {
			t.Errorf("the discard also emitted EventZoneMove — Syr Konrad counts it twice")
		}
	}
}

// --- (b) the effect discard: Mind Rot on a commander -----------------

// TestEffectDiscardOfACommanderOffersTheCommandZone is the issue, on
// the path most discards take: QueueDiscardChoiceForEffect's
// continuation (#797). Both answers, and the prompt's own "then" waits
// for the card to land.
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
				Then:   func(*Game) error { thenRuns++; return nil },
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

			// The CR 903.9 prompt gates the move: nothing has left the
			// hand, nothing has been emitted, and "then draw a card"
			// has not happened yet.
			if !p.Hand.Contains(cmdID) {
				t.Fatal("the commander left the hand before the prompt was answered")
			}
			if len(w.discards) != 0 {
				t.Errorf("EventDiscardCard fired before the card moved: %d", len(w.discards))
			}
			if thenRuns != 0 {
				t.Errorf(`"then" ran %d times before the discard landed`, thenRuns)
			}

			prompt := expectCommanderPrompt(t, g, p)
			if got := commanderPromptCard(t, prompt); got != cmdID {
				t.Fatalf("the prompt is about %s, want the discarded commander %s", got, cmdID)
			}
			if err := g.ResolveOptionalReplacement(prompt.ID, p.ID, toCommandZone); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}
			if len(g.PendingChoices) != 0 {
				t.Fatalf("%d prompts survived the answer", len(g.PendingChoices))
			}

			want, other := p.Command, p.Graveyard
			if !toCommandZone {
				want, other = p.Graveyard, p.Command
			}
			assertOnlyIn(t, cmdID, want, other, p.Hand, g.Exile)
			assertOneDiscardEvent(t, w, p.ID, cmdID, want.Kind)
			if thenRuns != 1 {
				t.Errorf(`"then" ran %d times, want 1 — once the discard had landed`, thenRuns)
			}
		})
	}
}

// --- (f) two commanders at once (partners) ---------------------------

// TestTwoDiscardedCommandersAskTwiceAndRunThenOnce — a partner pair
// pitched to one Mind Rot opens two windows. They are sequenced
// through the resume, one per card, and the prompt's own "then" runs
// once, after the LAST of them has landed.
func TestTwoDiscardedCommandersAskTwiceAndRunThenOnce(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	ids := emptyHandWithCommanders(t, g, p, 2)
	first, second := ids[0], ids[1]
	w := watchDiscards(g)

	thenRuns := 0
	queueEffectDiscard(g, DiscardPrompt{
		Player: p.ID,
		N:      2,
		Then:   func(*Game) error { thenRuns++; return nil },
	})
	c := discardPromptFor(g, p.ID)
	if c == nil {
		t.Fatal("no discard prompt queued")
	}
	if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{first, second}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	// One window at a time: the second commander is still in hand,
	// untouched, while the first one's question is open.
	promptA := expectCommanderPrompt(t, g, p)
	if got := commanderPromptCard(t, promptA); got != first {
		t.Fatalf("first prompt is about %s, want %s", got, first)
	}
	if !p.Hand.Contains(second) {
		t.Error("the second commander moved while the first one's prompt was open")
	}
	if err := g.ResolveOptionalReplacement(promptA.ID, p.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement (first): %v", err)
	}

	promptB := expectCommanderPrompt(t, g, p)
	if promptB.ID == promptA.ID {
		t.Fatal("answering the first prompt did not queue a second one")
	}
	if got := commanderPromptCard(t, promptB); got != second {
		t.Fatalf("second prompt is about %s, want %s", got, second)
	}
	if thenRuns != 0 {
		t.Errorf(`"then" ran %d times with half the batch still owed`, thenRuns)
	}
	if err := g.ResolveOptionalReplacement(promptB.ID, p.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement (second): %v", err)
	}

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the batch", len(g.PendingChoices))
	}
	assertOnlyIn(t, first, p.Command, p.Graveyard, p.Hand)
	assertOnlyIn(t, second, p.Graveyard, p.Command, p.Hand)
	if thenRuns != 1 {
		t.Errorf(`"then" ran %d times, want exactly 1 for the batch`, thenRuns)
	}

	// (h) one event per card, whichever way each one went — the
	// "whenever you discard a card" family fires twice, not once and
	// not three times.
	if len(w.discards) != 2 {
		t.Fatalf("EventDiscardCard x %d, want 2 (one per card)", len(w.discards))
	}
	landed := map[uuid.UUID]ZoneKind{}
	for _, ev := range w.discards {
		if ev.Actor != p.ID || ev.OldZone != ZoneHand {
			t.Errorf("EventDiscardCard actor/oldzone = %s/%s, want %s/hand", ev.Actor, ev.OldZone, p.ID)
		}
		landed[ev.CardID] = ev.NewZone
	}
	if landed[first] != ZoneCommand {
		t.Errorf("first commander's discard event says %s, want command", landed[first])
	}
	if landed[second] != ZoneGraveyard {
		t.Errorf("second commander's discard event says %s, want graveyard", landed[second])
	}
}

// --- (d) the random discard ------------------------------------------

// TestRandomDiscardOfACommanderOffersTheCommandZone — CR 701.8b picks
// the card, CR 903.9 still asks about it. A one-card hand makes the
// random pick the commander.
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
	if !p.Hand.Contains(cmdID) {
		t.Fatal("the commander left the hand before the prompt was answered")
	}

	prompt := expectCommanderPrompt(t, g, p)
	if err := g.ResolveOptionalReplacement(prompt.ID, p.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, p.Command, p.Graveyard, p.Hand)
	assertOneDiscardEvent(t, w, p.ID, cmdID, ZoneCommand)
}

// --- (c) the cleanup discard -----------------------------------------

// TestCleanupDiscardOfACommanderOffersTheCommandZone — the CR 514.1
// hand-size discard is a turn-based action, not an effect, and it asks
// too. What it additionally owes is its own bookkeeping: the pending
// entry drains and the cursor walks on only once the whole batch has
// landed, which is now on the far side of the prompt.
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
			if !active.Hand.Contains(cmdID) {
				t.Fatal("the commander left the hand before the prompt was answered")
			}
			if g.DiscardPending[active.ID] != 2 {
				t.Error("the hand-size debt was cleared before the discard landed")
			}

			prompt := expectCommanderPrompt(t, g, active)
			if err := g.ResolveOptionalReplacement(prompt.ID, active.ID, toCommandZone); err != nil {
				t.Fatalf("ResolveOptionalReplacement: %v", err)
			}

			want, other := active.Command, active.Graveyard
			if !toCommandZone {
				want, other = active.Graveyard, active.Command
			}
			assertOnlyIn(t, cmdID, want, other, active.Hand)
			assertOnlyIn(t, spare, active.Graveyard, active.Hand)
			if len(w.discards) != 2 {
				t.Errorf("EventDiscardCard x %d, want 2", len(w.discards))
			}
			// The cleanup finishes: the debt is paid and the cursor
			// has walked on to the next seat.
			if len(g.DiscardPending) != 0 {
				t.Errorf("DiscardPending survived the cleanup discard: %+v", g.DiscardPending)
			}
			if g.Turn.ActiveSeat == 0 {
				t.Error("the cursor stayed on seat 0 after the cleanup discard resolved")
			}
		})
	}
}

// --- (e) the cost discard, which may not ask --------------------------

// TestCostDiscardOfACommanderNeverAsks — CR 601.2h pays a spell's
// costs as one indivisible step, so this one discard must NOT pause.
// CR 903.9 is a "may" and a cost that cannot ask falls back to the
// ordinary result: the commander goes to the graveyard and the spell
// is cast.
func TestCostDiscardOfACommanderNeverAsks(t *testing.T) {
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
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a cost paused on %d prompt(s) — CR 601.2h pays costs as one step", len(g.PendingChoices))
	}
	assertOnlyIn(t, cmdID, me.Graveyard, me.Command, me.Hand)
	if !g.Stack.Contains(spell) {
		t.Error("the spell was not cast")
	}
	assertOneDiscardEvent(t, w, me.ID, cmdID, ZoneGraveyard)
}

// --- (g) undo across the open prompt ----------------------------------

// TestUndoAcrossADiscardPauseReplaysTheSameWay — the rest of a batch
// is held on an unserialisable continuation, so the undo stack has to
// be able to rewind a game sitting on one and have the replay come out
// the same.
//
// Two rewinds, because they fail differently. Rewinding to BEFORE the
// discard drops the prompt and the continuation together. Rewinding to
// WHILE THE PROMPT IS OPEN keeps them, and answering a second time has
// to discard the same two cards into the same two zones — which is why
// the batch is the remaining slice carried forward rather than a
// shared cursor the rewind could not put back.
func TestUndoAcrossADiscardPauseReplaysTheSameWay(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	cmdID := emptyHandWithCommanders(t, g, p, 1)[0]
	spare := NewCard("Spare", p.ID)
	spare.TypeLine = "Sorcery"
	g.WithWriteLock(func() { p.Hand.PushTop(spare) })

	thenRuns := 0
	pitch := func() {
		t.Helper()
		queueEffectDiscard(g, DiscardPrompt{
			Player: p.ID,
			N:      2,
			Then:   func(*Game) error { thenRuns++; return nil },
		})
		c := discardPromptFor(g, p.ID)
		if c == nil {
			t.Fatal("no discard prompt queued")
		}
		if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{cmdID, spare.InstanceID}); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
	}

	// --- rewind to before the discard ---
	beforeDiscard := g.Clone()
	pitch()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want the CR 903.9 prompt", len(g.PendingChoices))
	}
	g.WithWriteLock(func() { g.RestoreFrom(beforeDiscard) })
	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the rewind to before the discard", len(g.PendingChoices))
	}
	if !g.Seats[1].Hand.Contains(cmdID) || !g.Seats[1].Hand.Contains(spare.InstanceID) {
		t.Fatal("the rewind did not put both cards back in hand")
	}

	// --- rewind into the open prompt, then answer twice ---
	thenRuns = 0
	pitch()
	promptOpen := g.Clone()
	answerOnlyCommanderPrompt(t, g, g.Seats[1].ID, true)
	if thenRuns != 1 {
		t.Fatalf(`first run: "then" ran %d times, want 1`, thenRuns)
	}
	assertOnlyIn(t, cmdID, g.Seats[1].Command, g.Seats[1].Graveyard, g.Seats[1].Hand)
	assertOnlyIn(t, spare.InstanceID, g.Seats[1].Graveyard, g.Seats[1].Hand)

	thenRuns = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if !g.Seats[1].Hand.Contains(cmdID) || !g.Seats[1].Hand.Contains(spare.InstanceID) {
		t.Fatal("the rewind into the open prompt did not put both cards back in hand")
	}
	answerOnlyCommanderPrompt(t, g, g.Seats[1].ID, true)
	if thenRuns != 1 {
		t.Errorf(`replay: "then" ran %d times, want 1 — the same answer must run the batch once`, thenRuns)
	}
	assertOnlyIn(t, cmdID, g.Seats[1].Command, g.Seats[1].Graveyard, g.Seats[1].Hand)
	assertOnlyIn(t, spare.InstanceID, g.Seats[1].Graveyard, g.Seats[1].Hand)
}

// answerOnlyCommanderPrompt answers the single queued CR 903.9 prompt.
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
		Then:   func(*Game) error { thenRuns++; return nil },
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
