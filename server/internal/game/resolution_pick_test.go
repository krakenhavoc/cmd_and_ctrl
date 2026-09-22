package game

import (
	"testing"

	"github.com/google/uuid"
)

// resolution_pick_test.go — #1214. The three resolution-time picks as
// SHAPES rather than through the cards that use them: the asked seat
// answers and nobody else can, the continuation runs exactly once after
// the last leg, a withdrawn prompt settles its leg instead of stranding
// the card, undo puts the question back, and the answer reaches the log.
//
// FOUR SEATS throughout, for ADR 0060 Decision 5's reason: in a
// two-player game a departure ends the game, so none of CR 800.4a's
// consequences is observable and a two-seat version of the departure
// tests would pass for the wrong reason.

// permanentOf puts a plain permanent on the battlefield under a seat's
// control and returns its instance ID.
func permanentOf(g *Game, controller uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Bear",
		Owner:      controller,
		Controller: controller,
	})
	return id
}

// choicesOfKind returns every open prompt of a kind, in queue order.
func choicesOfKind(g *Game, kind PendingChoiceKind) []*PendingChoice {
	var out []*PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			out = append(out, c)
		}
	}
	return out
}

// --- reveal_pick -----------------------------------------------------

// TestRevealPickAsksTheNamedSeatAndPartitionsTheSet is the kind's core
// contract, and it is Intuition's sentence: the opponent picks, only
// they may answer, and the continuation is handed BOTH halves in reveal
// order.
func TestRevealPickAsksTheNamedSeatAndPartitionsTheSet(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Intuition")
	revealed := handCardOf(t, caster, 3)

	var picked, left []uuid.UUID
	runs := 0
	g.WithWriteLock(func() {
		n, err := g.RevealPickThenForEffect(RevealPickPrompt{
			Chooser:  opp.ID,
			Owner:    caster.ID,
			Source:   source,
			Question: "Intuition — choose one",
			Cards:    revealed,
			Min:      1,
			Max:      1,
		}, func(_ *Game, gotPicked, gotLeft []uuid.UUID) error {
			runs++
			picked, left = gotPicked, gotLeft
			return nil
		})
		if err != nil || n != 1 {
			t.Fatalf("RevealPickThenForEffect queued %d prompts, err %v", n, err)
		}
	})

	open := choicesOfKind(g, PendingChoiceRevealPick)
	if len(open) != 1 {
		t.Fatalf("want one reveal_pick, got %+v", g.PendingChoices)
	}
	c := open[0]
	if c.Chooser != opp.ID || c.FromPlayer != caster.ID {
		t.Errorf("the question goes to the opponent about the caster's cards: chooser %s, from %s", c.Chooser, c.FromPlayer)
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("bounds ride the prompt: %d..%d", c.ChooseMin, c.ChooseMax)
	}
	if !ChoiceBlocksTable(c.Kind) {
		t.Error("a resolution-time pick stops the table")
	}
	if err := g.ResolveRevealPick(c.ID, caster.ID, revealed[:1]); err == nil {
		t.Error("only the chooser may answer")
	}
	if err := g.ResolveRevealPick(c.ID, opp.ID, revealed[:2]); err == nil {
		t.Error("a pick outside the bounds is refused")
	}
	if len(g.PendingChoices) != 1 {
		t.Fatal("a refused answer leaves the prompt open")
	}

	if err := g.ResolveRevealPick(c.ID, opp.ID, revealed[1:2]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if runs != 1 {
		t.Errorf("the continuation ran %d times, want once", runs)
	}
	if len(picked) != 1 || picked[0] != revealed[1] {
		t.Errorf("picked %v, want %v", picked, revealed[1:2])
	}
	if len(left) != 2 || left[0] != revealed[0] || left[1] != revealed[2] {
		t.Errorf("left %v, want the other two in reveal order", left)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the prompt is dequeued: %+v", g.PendingChoices)
	}
}

// TestRevealPickRunsItsContinuationWithNothingAsked — the #544 / #1006
// contract at queue time: a chooser who has already left is not asked,
// and the rest of the card still resolves with nothing picked.
func TestRevealPickRunsItsContinuationWithNothingAsked(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, gone := g.Seats[0], g.Seats[1]
	revealed := handCardOf(t, caster, 3)
	if err := g.Concede(gone.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	var picked, left []uuid.UUID
	runs := 0
	g.WithWriteLock(func() {
		n, err := g.RevealPickThenForEffect(RevealPickPrompt{
			Chooser: gone.ID,
			Owner:   caster.ID,
			Cards:   revealed,
			Min:     1,
			Max:     1,
		}, func(_ *Game, gotPicked, gotLeft []uuid.UUID) error {
			runs++
			picked, left = gotPicked, gotLeft
			return nil
		})
		if err != nil {
			t.Fatalf("RevealPickThenForEffect: %v", err)
		}
		if n != 0 {
			t.Fatalf("queued %d prompts for a seat that has left", n)
		}
	})
	if runs != 1 {
		t.Fatalf("the continuation ran %d times, want once", runs)
	}
	if len(picked) != 0 || len(left) != 3 {
		t.Errorf("nothing picked, everything left: %v / %v", picked, left)
	}
}

// TestRevealPickIsReassignedWhenItsChooserLeaves — CR 800.4g's second
// sentence: the choice was to be made by an opponent of the object's
// controller, so another opponent makes it.
func TestRevealPickIsReassignedWhenItsChooserLeaves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver, next := g.Seats[0], g.Seats[1], g.Seats[2]
	source := departureTestSource(g, caster.ID, "Gifts Ungiven")
	revealed := handCardOf(t, caster, 3)

	var picked []uuid.UUID
	runs := 0
	g.WithWriteLock(func() {
		if _, err := g.RevealPickThenForEffect(RevealPickPrompt{
			Chooser:  leaver.ID,
			Owner:    caster.ID,
			Source:   source,
			Question: "Gifts Ungiven — choose two",
			Cards:    revealed,
			Min:      2,
			Max:      2,
		}, func(_ *Game, gotPicked, _ []uuid.UUID) error {
			runs++
			picked = gotPicked
			return nil
		}); err != nil {
			t.Fatalf("RevealPickThenForEffect: %v", err)
		}
	})
	id := choicesOfKind(g, PendingChoiceRevealPick)[0].ID

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := findChoice(g, id)
	if c == nil {
		t.Fatal("the pick was dropped; CR 800.4g reassigns a question about somebody else's cards")
	}
	if c.Chooser != next.ID {
		t.Fatalf("the pick went to %s, want seat 2 (%s)", c.Chooser, next.ID)
	}
	if len(c.ChooseCards) != 3 {
		t.Errorf("the candidates belong to the caster and survive: %d, want 3", len(c.ChooseCards))
	}
	if runs != 0 {
		t.Fatal("the continuation ran while the question was still open")
	}
	if err := g.ResolveRevealPick(id, next.ID, revealed[:2]); err != nil {
		t.Fatalf("the inheritor could not answer: %v", err)
	}
	if runs != 1 || len(picked) != 2 {
		t.Errorf("the card's continuation ran %d times with %v", runs, picked)
	}
}

// TestRevealPickUndoesAcrossThePrompt — rewinding past the answer puts
// the question back and the restored prompt answers the same way. The
// RUN has to rewind with the queue, which is what clonePromptRuns is
// for: a restored prompt whose run had already settled would answer
// into nothing.
func TestRevealPickUndoesAcrossThePrompt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	revealed := handCardOf(t, caster, 3)

	runs := 0
	var picked []uuid.UUID
	g.WithWriteLock(func() {
		if _, err := g.RevealPickThenForEffect(RevealPickPrompt{
			Chooser: opp.ID,
			Owner:   caster.ID,
			Cards:   revealed,
			Min:     1,
			Max:     1,
		}, func(_ *Game, gotPicked, _ []uuid.UUID) error {
			runs++
			picked = gotPicked
			return nil
		}); err != nil {
			t.Fatalf("RevealPickThenForEffect: %v", err)
		}
	})
	id := choicesOfKind(g, PendingChoiceRevealPick)[0].ID

	snap := g.Clone()
	if err := g.ResolveRevealPick(id, opp.ID, revealed[:1]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if runs != 1 {
		t.Fatalf("the continuation ran %d times before the undo", runs)
	}

	g.RestoreFrom(snap)
	runs, picked = 0, nil
	restored := choicesOfKind(g, PendingChoiceRevealPick)
	if len(restored) != 1 {
		t.Fatalf("undo puts the question back: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(restored[0].ID, opp.ID, revealed[2:3]); err != nil {
		t.Fatalf("ResolveRevealPick after undo: %v", err)
	}
	if runs != 1 || len(picked) != 1 || picked[0] != revealed[2] {
		t.Errorf("the restored run settles once, with %v", picked)
	}
}

// TestRevealPickNarratesTheAnswer — #1023's gate. The decision reaches
// the event log with the chooser's name on it, whatever the pick's size.
func TestRevealPickNarratesTheAnswer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, opp := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Intuition")
	revealed := handCardOf(t, caster, 3)
	before := countEvents(g, EventCardsChosen)

	g.WithWriteLock(func() {
		if _, err := g.RevealPickThenForEffect(RevealPickPrompt{
			Chooser: opp.ID,
			Owner:   caster.ID,
			Source:  source,
			Cards:   revealed,
			Min:     1,
			Max:     1,
		}, func(*Game, []uuid.UUID, []uuid.UUID) error { return nil }); err != nil {
			t.Fatalf("RevealPickThenForEffect: %v", err)
		}
	})
	id := choicesOfKind(g, PendingChoiceRevealPick)[0].ID
	if err := g.ResolveRevealPick(id, opp.ID, revealed[:1]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if countEvents(g, EventCardsChosen) != before+1 {
		t.Fatalf("the answer was not narrated: %d events of the kind", countEvents(g, EventCardsChosen))
	}
	var ev Event
	for _, e := range g.Events {
		if e.Kind == EventCardsChosen {
			ev = e
		}
	}
	if ev.Actor != opp.ID {
		t.Errorf("the line names %s as the chooser, want %s", ev.Actor, opp.ID)
	}
	if ev.Amount != 1 || ev.CardID != revealed[0] {
		t.Errorf("a one-card pick names the card: amount %d, card %s", ev.Amount, ev.CardID)
	}
	if ev.Source != source {
		t.Errorf("the line names the card that asked: %s, want %s", ev.Source, source)
	}
}

// --- their_permanents / own_permanents -------------------------------

// TestPermanentPickRunAsksPerSeatAndSettlesOnce is Tragic Arrogance's
// shape: ONE printed instruction, one prompt per player, one
// continuation after the last of them — and the kind is decided per leg
// by whose board is on offer.
func TestPermanentPickRunAsksPerSeatAndSettlesOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, them, third := g.Seats[0], g.Seats[1], g.Seats[2]
	source := departureTestSource(g, me.ID, "Tragic Arrogance")
	mine := permanentOf(g, me.ID, "My Bear")
	theirs := permanentOf(g, them.ID, "Their Bear")
	thirds := permanentOf(g, third.ID, "Third Bear")

	runs := 0
	var answer PromptedPicks
	g.WithWriteLock(func() {
		n, err := g.PermanentsPickedThenForEffect(PermanentPickPrompt{
			Chooser:  me.ID,
			Source:   source,
			Question: "Tragic Arrogance — keep one",
			Of:       []uuid.UUID{me.ID, them.ID, third.ID},
			Candidates: func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == of && c.IsCreature() {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
		}, func(_ *Game, picked PromptedPicks) error {
			runs++
			answer = picked
			return nil
		})
		if err != nil || n != 3 {
			t.Fatalf("queued %d prompts, err %v", n, err)
		}
	})

	own := choicesOfKind(g, PendingChoiceOwnPermanents)
	theirsPrompts := choicesOfKind(g, PendingChoiceTheirPermanents)
	if len(own) != 1 || len(theirsPrompts) != 2 {
		t.Fatalf("the leg over the chooser's own board is own_permanents and the rest are not: %d / %d", len(own), len(theirsPrompts))
	}
	for _, c := range theirsPrompts {
		if c.Chooser != me.ID || c.FromPlayer == me.ID {
			t.Errorf("a cross-table leg is asked of the chooser about somebody else: %s / %s", c.Chooser, c.FromPlayer)
		}
	}

	// Every leg is owed by the SAME seat, and nobody else may answer.
	first := own[0]
	if err := g.ResolveOwnPermanents(first.ID, them.ID, []uuid.UUID{mine}); err == nil {
		t.Error("only the chooser may answer")
	}
	// A prompt may not be answered through the other kind's verb.
	if err := g.ResolveTheirPermanents(first.ID, me.ID, []uuid.UUID{mine}); err == nil {
		t.Error("own_permanents is not answerable as their_permanents")
	}

	if err := g.ResolveOwnPermanents(first.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if runs != 0 {
		t.Fatal("the run settled before its last leg")
	}
	for _, c := range theirsPrompts {
		want := theirs
		if c.FromPlayer == third.ID {
			want = thirds
		}
		if err := g.ResolveTheirPermanents(c.ID, me.ID, []uuid.UUID{want}); err != nil {
			t.Fatalf("ResolveTheirPermanents: %v", err)
		}
	}
	if runs != 1 {
		t.Fatalf("the continuation ran %d times, want once", runs)
	}
	if answer.Count() != 3 {
		t.Errorf("three permanents chosen across the run, got %d", answer.Count())
	}
	if !answer.Contains(theirs) || !answer.Picked(third.ID) {
		t.Errorf("the answer is keyed by WHOSE permanents were chosen: %+v", answer)
	}
	if got := answer.By(them.ID); len(got) != 1 || got[0] != theirs {
		t.Errorf("seat 1's entry is their own permanent: %v", got)
	}
}

// TestPermanentPickSkipsASeatWithNothingToOffer — CR 608.2's "as much
// as possible" at queue time: a seat with no candidate is asked nothing
// and has no entry, rather than being handed an empty question.
func TestPermanentPickSkipsASeatWithNothingToOffer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, empty := g.Seats[0], g.Seats[1]
	mine := permanentOf(g, me.ID, "My Bear")

	runs := 0
	var answer PromptedPicks
	g.WithWriteLock(func() {
		n, err := g.PermanentsPickedThenForEffect(PermanentPickPrompt{
			Chooser: me.ID,
			Of:      []uuid.UUID{me.ID, empty.ID},
			Candidates: func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == of && c.IsCreature() {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
		}, func(_ *Game, picked PromptedPicks) error {
			runs++
			answer = picked
			return nil
		})
		if err != nil || n != 1 {
			t.Fatalf("queued %d prompts, err %v", n, err)
		}
	})
	id := choicesOfKind(g, PendingChoiceOwnPermanents)[0].ID
	if err := g.ResolveOwnPermanents(id, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if runs != 1 {
		t.Fatalf("the continuation ran %d times, want once", runs)
	}
	if len(answer) != 1 || answer.Picked(empty.ID) {
		t.Errorf("the skipped seat has no entry: %+v", answer)
	}
}

// TestOwnPermanentsPickAcceptsAnEmptyAnswer — "sacrifice ANY NUMBER of
// lands" (Scapeshift). A floor of zero is a real answer and the
// continuation reads the count.
func TestOwnPermanentsPickAcceptsAnEmptyAnswer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	permanentOf(g, me.ID, "My Bear")

	var answer PromptedPicks
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenForEffect(PermanentPickPrompt{
			Chooser: me.ID,
			Of:      []uuid.UUID{me.ID},
			Candidates: func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == of {
						out = append(out, c.InstanceID)
					}
				}
				return out, 0, 0
			},
		}, func(_ *Game, picked PromptedPicks) error {
			answer = picked
			return nil
		}); err != nil {
			t.Fatalf("PermanentsPickedThenForEffect: %v", err)
		}
	})
	c := choicesOfKind(g, PendingChoiceOwnPermanents)[0]
	if c.ChooseMin != 0 {
		t.Fatalf("an any-number pick has a floor of zero, got %d", c.ChooseMin)
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveOwnPermanents(empty): %v", err)
	}
	if answer.Count() != 0 {
		t.Errorf("nothing was chosen: %+v", answer)
	}
}

// TestOwnPermanentsPickIsDroppedAndSettlesItsLeg — the departure table's
// two columns at once. own_permanents is never reassigned (the pool is
// the departed seat's own board, CR 800.4a), and its drop still settles
// the leg so the rest of the card runs (#1006's dropDefault).
func TestOwnPermanentsPickIsDroppedAndSettlesItsLeg(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Bane of Bala Ged")
	permanentOf(g, leaver.ID, "Their Bear")

	runs := 0
	var answer PromptedPicks
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenForEffect(PermanentPickPrompt{
			Chooser: leaver.ID,
			Source:  source,
			Of:      []uuid.UUID{leaver.ID},
			Candidates: func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == of {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
		}, func(_ *Game, picked PromptedPicks) error {
			runs++
			answer = picked
			return nil
		}); err != nil {
			t.Fatalf("PermanentsPickedThenForEffect: %v", err)
		}
	})
	id := choicesOfKind(g, PendingChoiceOwnPermanents)[0].ID

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if findChoice(g, id) != nil {
		t.Error("the prompt is about the departed seat's own board and is dropped, not reassigned")
	}
	if runs != 1 {
		t.Fatalf("the withdrawn leg did not settle: the continuation ran %d times", runs)
	}
	if answer.Count() != 0 {
		t.Errorf("a withdrawn leg chooses nothing on the departed seat's behalf: %+v", answer)
	}
}

// TestTheirPermanentsPickIsReassigned — the other column. The prompt is
// about a living seat's board, so CR 800.4g moves it.
func TestTheirPermanentsPickIsReassigned(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, leaver, third := g.Seats[0], g.Seats[1], g.Seats[2]
	source := departureTestSource(g, me.ID, "Tragic Arrogance")
	bear := permanentOf(g, third.ID, "Third Bear")

	runs := 0
	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenForEffect(PermanentPickPrompt{
			Chooser: leaver.ID,
			Source:  source,
			Of:      []uuid.UUID{third.ID},
			Candidates: func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == of {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
		}, func(*Game, PromptedPicks) error {
			runs++
			return nil
		}); err != nil {
			t.Fatalf("PermanentsPickedThenForEffect: %v", err)
		}
	})
	id := choicesOfKind(g, PendingChoiceTheirPermanents)[0].ID

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := findChoice(g, id)
	if c == nil {
		t.Fatal("the prompt was dropped; it is about a living seat's board (CR 800.4g)")
	}
	if c.Chooser == leaver.ID {
		t.Fatal("the prompt is still addressed to the seat that left")
	}
	if runs != 0 {
		t.Fatal("the continuation ran while the question was still open")
	}
	if err := g.ResolveTheirPermanents(id, c.Chooser, []uuid.UUID{bear}); err != nil {
		t.Fatalf("the inheritor could not answer: %v", err)
	}
	if runs != 1 {
		t.Errorf("the continuation ran %d times, want once", runs)
	}
}

// TestPermanentPickPrunesAPermanentThatLeavesUnderIt — the live-zone
// re-check, inherited from the choose-cards payload: a permanent that
// leaves the battlefield while the question hangs comes off the prompt,
// so the enumerator never offers a set the resolver would refuse.
func TestPermanentPickPrunesAPermanentThatLeavesUnderIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	keep := permanentOf(g, me.ID, "Keeper")
	doomed := permanentOf(g, me.ID, "Doomed")

	g.WithWriteLock(func() {
		if _, err := g.PermanentsPickedThenForEffect(PermanentPickPrompt{
			Chooser: me.ID,
			Of:      []uuid.UUID{me.ID},
			Candidates: func(g *Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == of {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
		}, func(*Game, PromptedPicks) error { return nil }); err != nil {
			t.Fatalf("PermanentsPickedThenForEffect: %v", err)
		}
	})
	c := choicesOfKind(g, PendingChoiceOwnPermanents)[0]
	if len(c.ChooseCards) != 2 {
		t.Fatalf("setup: %d candidates, want 2", len(c.ChooseCards))
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(doomed); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
		g.pruneCardSetChoicesLocked()
	})
	c = choicesOfKind(g, PendingChoiceOwnPermanents)[0]
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != keep {
		t.Fatalf("the departed permanent is still offered: %v", c.ChooseCards)
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{doomed}); err == nil {
		t.Error("the resolver refuses a permanent that has left")
	}
}
