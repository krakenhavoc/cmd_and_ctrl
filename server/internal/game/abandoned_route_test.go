package game

import (
	"testing"

	"github.com/google/uuid"
)

// abandoned_route_test.go pins #865: a paused exit whose prompt is
// taken AWAY rather than answered still reaches its route's
// continuation, so a batch sequenced through that continuation carries
// on instead of stalling.
//
// Two things take a prompt away, and before the fix both simply
// discarded the frame:
//
//   - DROPPED — the chooser left the game, so the question can never be
//     answered (cleanupStackForEliminatedLocked, CR 800.4a);
//   - PRUNED — the card moved by some other route while the question
//     was open, so the move it asks about can never happen
//     (pruneStaleZoneChangeChoicesLocked, #605/#701).
//
// The continuation is how a multi-card discard (#853) and a sequenced
// wipe (#815) carry the rest of the batch and the caller's own "then":
// a two-card Mind Rot stopped after the first card and a board wipe
// stopped after the commander, with the "for each creature destroyed
// this way" clause never running at all. Same terminal-outcome shape
// #808 gave the life and damage tails when a player leaves.
//
// The abandoned leg is NOT LANDED: nothing moved, so it is not counted
// as destroyed or discarded and no EventDiscardCard is emitted for it —
// CR 701.8a defines a discard as the move out of the hand, and there
// was none.

// wipeCommanderOf puts a commander on the battlefield under `owner`.
func wipeCommanderOf(t *testing.T, g *Game, owner *Player) uuid.UUID {
	t.Helper()
	return seatCommander(t, g.Battlefield, owner)
}

// --- (a) DROPPED: the chooser concedes ------------------------------

// TestADroppedDiscardPromptStillRunsTheRestOfTheBatch is the discard
// half of the issue. Two commanders are pitched to one Mind Rot; their
// owner concedes with the first card's CR 903.9 prompt open. The
// second card is still owed, and so is the prompt's own "then".
//
// The second leg does not ask a question of its own: its chooser is the
// same player, and a "may" whose chooser has left the game is declined
// rather than queued (offerOptionalReplacementLocked). What matters is
// that the batch reaches it at all.
func TestADroppedDiscardPromptStillRunsTheRestOfTheBatch(t *testing.T) {
	g := newFourPlayerActiveGame(t)
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
	prompt := expectCommanderPrompt(t, g, p)
	if got := commanderPromptCard(t, prompt); got != first {
		t.Fatalf("the open prompt is about %s, want the first card %s", got, first)
	}

	if err := g.Concede(p.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the concede", len(g.PendingChoices))
	}
	if thenRuns != 1 {
		t.Errorf(`the prompt's "then" ran %d times, want exactly 1 — the batch owes it once, `+
			`whether the last leg landed or was abandoned`, thenRuns)
	}

	// Where the two cards ENDED is no longer observable: CR 800.4a
	// (#769) takes everything the conceding player owns out of the game
	// a moment after the batch finishes, so "still in hand" and "in the
	// graveyard" are both "gone". What the batch did is still pinned,
	// and by the thing that was always the real claim: the events. The
	// abandoned leg moved nothing, so it discarded nothing, and the
	// "whenever you discard a card" family sees one card, not two.
	for what, id := range map[string]uuid.UUID{"first": first, "second": second} {
		if zone := zoneHoldingCard(g, id); zone != "" {
			t.Errorf("the %s card is in %s: it should have left the game with its owner", what, zone)
		}
	}
	if len(w.discards) != 1 || w.discards[0].CardID != second {
		t.Errorf("EventDiscardCard x %d (%v), want exactly one, for the card that actually left the hand",
			len(w.discards), w.discards)
	}
}

// TestADroppedWipePromptStillFinishesTheSweep is the destroy half, and
// the shape that shows a second prompt surviving: two players'
// commanders are caught in one wipe, and the first owner concedes while
// their prompt is open. The second owner is still at the table, so
// their question appears; the sweep and the "for each creature
// destroyed this way" clause both finish.
func TestADroppedWipePromptStillFinishesTheSweep(t *testing.T) {
	// Four seats: the sweep must still have a table to finish on after
	// one of them concedes.
	g := newFourPlayerActiveGame(t)
	leaver, stayer := g.Seats[1], g.Seats[2]
	gone := wipeCommanderOf(t, g, leaver)
	asked := wipeCommanderOf(t, g, stayer)
	bear := wipeTarget(g, g.Seats[0])

	got, ran := destroyAllThen(t, g, []uuid.UUID{gone, asked, bear})
	if *ran != 0 {
		t.Fatalf("the continuation ran %d times with the first prompt open, want 0", *ran)
	}
	first := expectCommanderPrompt(t, g, leaver)
	if card := commanderPromptCard(t, first); card != gone {
		t.Fatalf("the open prompt is about %s, want %s", card, gone)
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	second := expectCommanderPrompt(t, g, stayer)
	if card := commanderPromptCard(t, second); card != asked {
		t.Fatalf("after the concede the open prompt is about %s, want the next commander %s (#865)", card, asked)
	}
	if *ran != 0 {
		t.Fatalf("the continuation ran %d times with the second prompt open, want 0", *ran)
	}
	if err := g.ResolveOptionalReplacement(second.ID, stayer.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1 for the sweep", *ran)
	}
	if !idsEqual(*got, []uuid.UUID{asked, bear}) {
		t.Errorf("destroyed this way = %v, want [asked bear] = %v %v — the abandoned leg destroyed nothing",
			*got, asked, bear)
	}
	// The abandoned leg destroyed nothing — pinned by `got` above. Its
	// commander is not on the battlefield either, but that is CR 800.4a
	// (#769) taking its owner's objects out of the game, not the route
	// moving it: it is in no zone at all, where a completed destruction
	// would have put it in a graveyard or a command zone.
	if zone := zoneHoldingCard(g, gone); zone != "" {
		t.Errorf("the departed player's commander is in %s, want out of the game entirely", zone)
	}
	if !stayer.Command.Contains(asked) {
		t.Error("the answered commander is in its owner's command zone")
	}
	if !g.Seats[0].Graveyard.Contains(bear) {
		t.Error("the rest of the sweep ran")
	}
}

// TestADroppedPromptInASingleCardRouteStillRunsThen — the batch is not
// the only caller waiting. A one-card sweep owes its continuation the
// same answer, and the answer is "nothing landed".
func TestADroppedPromptInASingleCardRouteStillRunsThen(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	commander := wipeCommanderOf(t, g, leaver)

	got, ran := destroyAllThen(t, g, []uuid.UUID{commander})
	if *ran != 0 {
		t.Fatalf("the continuation ran %d times with the prompt open, want 0", *ran)
	}
	expectCommanderPrompt(t, g, leaver)

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if *ran != 1 {
		t.Fatalf("the continuation ran %d times after the drop, want exactly 1 (#865)", *ran)
	}
	if len(*got) != 0 {
		t.Errorf("destroyed this way = %v, want none — the destruction never happened", *got)
	}
	// Nothing moved: no graveyard, no command zone, no exile. The
	// commander is in no zone at all, because CR 800.4a (#769) took it
	// out of the game with its owner — which is not the route landing.
	if zone := zoneHoldingCard(g, commander); zone != "" {
		t.Errorf("the commander is in %s: the abandoned route moved it after all", zone)
	}
}

// --- (b) PRUNED: the card leaves by another route --------------------

// TestAPrunedWipePromptStillFinishesTheSweep is the prune half. A
// commander caught in a wipe stops to answer CR 903.9; while the
// question is open the card leaves the battlefield by another route
// (flickered out from under the wipe), which makes the prompt
// unanswerable — the move it asks about can never happen — and #605
// withdraws it. The rest of the sweep and the "for each" clause are
// still owed, and the card that left is not counted: it was never put
// into a graveyard, so by CR 701.7a it was not destroyed.
func TestAPrunedWipePromptStillFinishesTheSweep(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	commander := wipeCommanderOf(t, g, owner)
	bear := wipeTarget(g, owner)

	got, ran := destroyAllThen(t, g, []uuid.UUID{commander, bear})
	expectCommanderPrompt(t, g, owner)
	if findBattlefieldCard(g, bear) == nil {
		t.Fatal("the rest of the sweep waits for the paused leg")
	}

	flickerToExile(t, g, commander)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("%d prompts survived the card leaving by another route", len(g.PendingChoices))
	}
	if *ran != 1 {
		t.Fatalf("the continuation ran %d times after the prune, want exactly 1 (#865)", *ran)
	}
	if !idsEqual(*got, []uuid.UUID{bear}) {
		t.Errorf("destroyed this way = %v, want just %v — a permanent that left to exile was never "+
			"put into a graveyard (CR 701.7a)", *got, bear)
	}
	if !g.Exile.Contains(commander) {
		t.Error("the flickered commander is in exile")
	}
	if !owner.Graveyard.Contains(bear) {
		t.Error("the rest of the sweep ran")
	}
}

// flickerToExile moves a card out from under an open prompt by another
// route. MustSettleNow keeps this exit from asking its own CR 903.9
// question (CR 601.2h's cost escape), so the only prompt in play stays
// the one under test.
func flickerToExile(t *testing.T, g *Game, cardID uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:        cardID,
			Dst:           ZoneExile,
			MustSettleNow: true,
		}); err != nil {
			t.Fatalf("routeCardToZoneLocked: %v", err)
		}
	})
}

// --- (c) undo across each -------------------------------------------

// TestUndoAcrossADroppedWipePromptReplays — rewinding into the open
// prompt and dropping it again has to sweep the same board and report
// the same list. The landed list is carried forward by value, so a
// replayed abandonment cannot see the first run's entries.
func TestUndoAcrossADroppedWipePromptReplays(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	gone := wipeCommanderOf(t, g, leaver)
	bear := wipeTarget(g, g.Seats[0])

	got, ran := destroyAllThen(t, g, []uuid.UUID{gone, bear})
	promptOpen := g.Clone()

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	first := append([]uuid.UUID(nil), (*got)...)
	if !idsEqual(first, []uuid.UUID{bear}) {
		t.Fatalf("first run destroyed %v, want just the bear", first)
	}

	*ran = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	if findBattlefieldCard(g, bear) == nil || findBattlefieldCard(g, gone) == nil {
		t.Fatal("the rewind puts both permanents back on the battlefield")
	}
	leaver = g.Seats[1]
	if leaver.Eliminated {
		t.Fatal("the rewind puts the conceded player back at the table")
	}
	expectCommanderPrompt(t, g, leaver)

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede (replay): %v", err)
	}
	if *ran != 1 {
		t.Errorf("the continuation ran %d times on the replay, want 1", *ran)
	}
	if !idsEqual(*got, first) {
		t.Errorf("replayed destroyed this way = %v, want the first run's %v", *got, first)
	}
}

// TestUndoAcrossAPrunedWipePromptReplays — the same contract for the
// prune.
func TestUndoAcrossAPrunedWipePromptReplays(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	commander := wipeCommanderOf(t, g, owner)
	bear := wipeTarget(g, owner)

	got, ran := destroyAllThen(t, g, []uuid.UUID{commander, bear})
	promptOpen := g.Clone()

	flickerToExile(t, g, commander)
	first := append([]uuid.UUID(nil), (*got)...)
	if !idsEqual(first, []uuid.UUID{bear}) {
		t.Fatalf("first run destroyed %v, want just the bear", first)
	}

	*ran = 0
	g.WithWriteLock(func() { g.RestoreFrom(promptOpen) })
	owner = g.Seats[0]
	if findBattlefieldCard(g, bear) == nil || findBattlefieldCard(g, commander) == nil {
		t.Fatal("the rewind puts both permanents back on the battlefield")
	}
	expectCommanderPrompt(t, g, owner)

	flickerToExile(t, g, commander)
	if *ran != 1 {
		t.Errorf("the continuation ran %d times on the replay, want 1", *ran)
	}
	if !idsEqual(*got, first) {
		t.Errorf("replayed destroyed this way = %v, want the first run's %v", *got, first)
	}
}
