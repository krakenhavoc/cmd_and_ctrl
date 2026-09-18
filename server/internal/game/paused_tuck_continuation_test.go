package game

import (
	"testing"

	"github.com/google/uuid"
)

// paused_tuck_continuation_test.go — #783. A library is a CR 903.9
// destination, so every tuck can PAUSE; this file pins the engine half
// of the answer, the continuation TuckToLibraryThenForEffect hands over
// instead of letting its caller write the next line.
//
// The contract is the one the exile batch signs (ADR 0013 §5k, §5n):
// the continuation runs EXACTLY ONCE, from every terminal outcome, with
// the card already where it ended up — and `tucked` is CR 400.7's
// reading, the card that ARRIVED in a library.

// tuckThen runs a tuck with a continuation that records what it saw.
// Returns pointers the test reads after the answer arrives.
func tuckThen(t *testing.T, g *Game, cardID uuid.UUID, opts TuckOptions) (runs *int, sawTucked *bool, sawZone *ZoneKind) {
	t.Helper()
	runs, sawTucked, sawZone = new(int), new(bool), new(ZoneKind)
	g.mu.Lock()
	err := g.TuckToLibraryThenForEffect(cardID, opts, func(g *Game, tucked bool) error {
		*runs++
		*sawTucked = tucked
		if z := g.findCardZoneLocked(cardID); z != nil {
			*sawZone = z.Kind
		}
		return nil
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("TuckToLibraryThenForEffect: %v", err)
	}
	return runs, sawTucked, sawZone
}

// TestTuckContinuationWaitsForTheCommandZoneAnswer is the bug. Nothing
// after the tuck may happen while the CR 903.9 question is open.
func TestTuckContinuationWaitsForTheCommandZoneAnswer(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	runs, _, _ := tuckThen(t, g, cmdID, TuckOptions{})
	if *runs != 0 {
		t.Fatalf("the continuation ran %d times with the prompt still open", *runs)
	}
	expectCommanderPrompt(t, g, owner)
}

// TestTuckContinuationAfterADeclineSeesTheCardInTheLibrary — the
// declining half: one run, the card in the library, `tucked` true.
func TestTuckContinuationAfterADeclineSeesTheCardInTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	runs, tucked, zone := tuckThen(t, g, cmdID, TuckOptions{})
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if *runs != 1 {
		t.Fatalf("continuation ran %d times, want exactly 1", *runs)
	}
	if !*tucked {
		t.Error("a declined commander went to the library, so it was tucked this way")
	}
	if *zone != ZoneLibrary {
		t.Errorf("the continuation saw the card in %s, want the library", *zone)
	}
	assertOnlyIn(t, cmdID, owner.Library, owner.Command, g.Battlefield)
}

// TestTuckContinuationAfterAnAcceptSeesTheCardInTheCommandZone — the
// accepting half. The continuation still runs (Chaos Warp still
// shuffles), but the card did not go THIS way (CR 400.7).
func TestTuckContinuationAfterAnAcceptSeesTheCardInTheCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	runs, tucked, zone := tuckThen(t, g, cmdID, TuckOptions{})
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if *runs != 1 {
		t.Fatalf("continuation ran %d times, want exactly 1", *runs)
	}
	if *tucked {
		t.Error("a commander that took the command zone was never put into a library (CR 400.7)")
	}
	if *zone != ZoneCommand {
		t.Errorf("the continuation saw the card in %s, want the command zone", *zone)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Library, g.Battlefield)
}

// TestUncontestedTuckRunsItsContinuationInline — a non-commander tuck
// is unchanged: nothing pauses, and the continuation has already run
// by the time the call returns.
func TestUncontestedTuckRunsItsContinuationInline(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
	})

	runs, tucked, zone := tuckThen(t, g, id, TuckOptions{})
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a non-commander tuck queued %d prompts", len(g.PendingChoices))
	}
	if *runs != 1 || !*tucked || *zone != ZoneLibrary {
		t.Errorf("inline tuck: runs=%d tucked=%v zone=%s; want 1/true/library", *runs, *tucked, *zone)
	}
	assertOnlyIn(t, id, owner.Library, g.Battlefield)
}

// TestTuckAtDepthLandsPositionedAfterADecline pins the God-Eternals'
// half of #783: "third from the top" is a POSITIONED LANDING carried on
// the route, not a remove-and-reinsert on the next line. The card is
// placed once, where the card says, on the far side of the prompt.
func TestTuckAtDepthLandsPositionedAfterADecline(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	if owner.Library.Size() < 3 {
		t.Fatalf("test needs a library of at least 3, got %d", owner.Library.Size())
	}
	cmdID := seatCommander(t, owner.Graveyard, owner)

	g.mu.Lock()
	err := g.TuckToLibraryAtDepthForEffect(cmdID, 3)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("TuckToLibraryAtDepthForEffect: %v", err)
	}
	if owner.Library.Contains(cmdID) {
		t.Fatal("the card moved while the CR 903.9 prompt was open")
	}
	for _, ev := range g.Events {
		if ev.Kind == EventEffectError {
			t.Fatalf("the tuck logged an effect error: %s", ev.ErrorMsg)
		}
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Library, owner.Command, owner.Graveyard)
	// Top is the last element; third from the top is two below it.
	at := len(owner.Library.Cards) - 3
	if at < 0 || owner.Library.Cards[at].InstanceID != cmdID {
		t.Error("a declined commander landed somewhere other than third from the top")
	}
	for _, ev := range g.Events {
		if ev.Kind == EventEffectError {
			t.Errorf("the resumed tuck logged an effect error: %s", ev.ErrorMsg)
		}
	}
}

// TestTuckBatchReportsOnlyTheLegsThatLanded — the batch form, and the
// reading Aetherspouts counts by: a commander that took the command
// zone is not among the cards its owner then arranges.
func TestTuckBatchReportsOnlyTheLegsThatLanded(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)
	plain := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: plain, Name: "Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
	})

	var landed []uuid.UUID
	runs := 0
	g.mu.Lock()
	err := g.TuckCardsToLibraryThenForEffect([]uuid.UUID{cmdID, plain}, TuckOptions{}, func(_ *Game, tucked []uuid.UUID) error {
		runs++
		landed = tucked
		return nil
	})
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("TuckCardsToLibraryThenForEffect: %v", err)
	}
	if runs != 0 {
		t.Fatal("the batch finished while the first leg's prompt was open")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	if runs != 1 {
		t.Fatalf("continuation ran %d times, want exactly 1", runs)
	}
	if len(landed) != 1 || landed[0] != plain {
		t.Errorf("landed = %v, want only the non-commander leg", landed)
	}
}
