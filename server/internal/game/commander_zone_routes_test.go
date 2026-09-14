package game

import (
	"testing"

	"github.com/google/uuid"
)

// commander_zone_routes_test.go pins CR 903.9 on the routes #529
// found unwired: counter, fizzle, exile, bounce, tuck and mill.
//
// The rule is "from ANYWHERE" — a commander that would be put into a
// library, hand, graveyard or exile from any zone gives its owner the
// command-zone choice. Before #529 the replacement window was opened
// by two callers only (MoveCardByIDAsCommander and
// routeBattlefieldCardToOwnerGraveyardLocked), both battlefield →
// graveyard, so every other way a commander leaves — which is most of
// how a commander actually leaves in a real game — moved the card raw
// and never prompted. #364 (countered) and #372 (exiled) are the two
// in-app reports that shape produced.
//
// Every test here asserts BOTH halves: the prompt is offered, AND the
// card lands where the answer says it should. Asserting only that the
// card left its old zone is what let this ship — see the note on
// TestWashAwayHardCastAnswersACommanderCast.

// seatCommander puts a commander card into `zone` and returns its ID.
func seatCommander(t *testing.T, zone *Zone, owner *Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	zone.PushTop(Card{
		InstanceID:  id,
		Name:        "Atraxa",
		TypeLine:    "Legendary Creature — Angel",
		Power:       4,
		Toughness:   4,
		Owner:       owner.ID,
		Controller:  owner.ID,
		IsCommander: true,
	})
	return id
}

// pushStackSpell puts a spell card onto the stack with a matching
// StackMeta entry, the shape CastSpell leaves behind.
func pushStackSpell(t *testing.T, g *Game, c Card) {
	t.Helper()
	g.Stack.PushTop(c)
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	g.StackMeta[c.InstanceID] = &StackItem{
		ID: c.InstanceID, Kind: StackItemSpell,
		Controller: c.Controller, Owner: c.Owner, SourceCardID: c.InstanceID,
	}
}

// expectCommanderPrompt asserts exactly one CR 903.9 optional
// replacement prompt is queued for `owner` and returns it.
func expectCommanderPrompt(t *testing.T, g *Game, owner *Player) *PendingChoice {
	t.Helper()
	if len(g.PendingChoices) != 1 {
		t.Fatalf("expected one CR 903.9 prompt, got %d pending choices", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]
	if prompt.Kind != PendingChoiceOptionalReplacement {
		t.Fatalf("prompt kind = %q, want %q", prompt.Kind, PendingChoiceOptionalReplacement)
	}
	if prompt.Chooser != owner.ID {
		t.Fatalf("chooser = %s, want commander owner %s", prompt.Chooser, owner.ID)
	}
	return prompt
}

// assertOnlyIn asserts the card is in `want` and in none of `others`.
func assertOnlyIn(t *testing.T, cardID uuid.UUID, want *Zone, others ...*Zone) {
	t.Helper()
	if !want.Contains(cardID) {
		t.Errorf("card not in %s as expected", want.Kind)
	}
	for _, z := range others {
		if z.Contains(cardID) {
			t.Errorf("card leaked into %s", z.Kind)
		}
	}
}

// --- counter (#364) ---------------------------------------------

// TestCommanderCounteredOffersCommandZone is #364: Wash Away counters
// a commander cast. Countering moves the spell stack → graveyard,
// which is a CR 903.9 destination, so the owner must be asked.
func TestCommanderCounteredOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Stack, owner)
	top, _ := g.Stack.Top()
	_, _ = g.Stack.Remove(cmdID)
	pushStackSpell(t, g, top)

	if err := g.CounterSpell(cmdID, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	// Nothing has moved: the prompt gates the move.
	if owner.Graveyard.Contains(cmdID) {
		t.Fatalf("countered commander hit the graveyard before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Graveyard, g.Stack, g.Exile)
	if _, ok := g.StackMeta[cmdID]; ok {
		t.Error("StackMeta entry survived the counter")
	}
}

// TestCommanderCounteredDeclineGoesToGraveyard — the owner may say no,
// and then the ordinary destination stands.
func TestCommanderCounteredDeclineGoesToGraveyard(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Stack, owner)
	top, _ := g.Stack.Top()
	_, _ = g.Stack.Remove(cmdID)
	pushStackSpell(t, g, top)

	if err := g.CounterSpell(cmdID, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Graveyard, owner.Command, g.Stack)
}

// TestNonCommanderCounterIsUnchanged — the pipeline must stay out of
// the way of the ~everything that is not a commander: no prompt, and
// the spell is in the graveyard when CounterSpell returns.
func TestNonCommanderCounterIsUnchanged(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	spellID := uuid.New()
	pushStackSpell(t, g, Card{
		InstanceID: spellID, Name: "Shock", TypeLine: "Instant",
		Owner: owner.ID, Controller: owner.ID,
	})
	if err := g.CounterSpell(spellID, nil); err != nil {
		t.Fatalf("CounterSpell: %v", err)
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("non-commander counter queued %d prompts, want 0", len(g.PendingChoices))
	}
	assertOnlyIn(t, spellID, owner.Graveyard, g.Stack)
}

// --- fizzle -----------------------------------------------------

// TestCommanderFizzledOffersCommandZone — "countered by game rules"
// (CR 608.2b, every target illegal on resolution) routes the spell
// off the stack through routeStackCardToGraveyardLocked, the same
// function the ordinary instant/sorcery resolution uses. A commander
// spell that fizzles is put into a graveyard, so 903.9 applies.
func TestCommanderFizzledOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Stack, owner)
	card, _ := g.Stack.Top()

	g.mu.Lock()
	err := g.routeStackCardToGraveyardLocked(card, "")
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("routeStackCardToGraveyardLocked: %v", err)
	}
	if owner.Graveyard.Contains(cmdID) {
		t.Fatalf("fizzled commander hit the graveyard before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Graveyard, g.Stack)
}

// --- exile (#372) -----------------------------------------------

// TestCommanderExiledOffersCommandZone is #372: Airbend exiles a
// commander off the battlefield. Exile is a CR 903.9 destination.
func TestCommanderExiledOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.ExileCardForEffect(cmdID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	if g.Exile.Contains(cmdID) {
		t.Fatalf("exiled commander hit exile before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, g.Exile, g.Battlefield)
}

// TestCommanderExiledDeclineGoesToExile — declining still exiles, and
// the LTB the exile owes still fires.
func TestCommanderExiledDeclineGoesToExile(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.ExileCardForEffect(cmdID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, g.Exile, owner.Command, g.Battlefield)
	if !hasEventFor(g, EventLTB, cmdID) {
		t.Error("declining the command zone skipped the leaves-the-battlefield event")
	}
}

// --- bounce -----------------------------------------------------

// TestCommanderBouncedOffersCommandZone — Unsummon on a commander.
// Hand is a CR 903.9 destination.
func TestCommanderBouncedOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.BounceToHandForEffect(cmdID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("BounceToHandForEffect: %v", err)
	}
	if owner.Hand.Contains(cmdID) {
		t.Fatalf("bounced commander hit the hand before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Hand, g.Battlefield)
}

// TestCommanderBouncedDeclineGoesToHand — and declining still bounces.
func TestCommanderBouncedDeclineGoesToHand(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.BounceToHandForEffect(cmdID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("BounceToHandForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Hand, owner.Command, g.Battlefield)
}

// --- tuck -------------------------------------------------------

// TestCommanderTuckedOffersCommandZone — Hinder / Condemn put a
// commander into a library, a CR 903.9 destination.
func TestCommanderTuckedOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.TuckToLibraryForEffect(cmdID, false)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("TuckToLibraryForEffect: %v", err)
	}
	if owner.Library.Contains(cmdID) {
		t.Fatalf("tucked commander hit the library before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Library, g.Battlefield)
}

// TestCommanderTuckedToBottomDeclineKeepsTheBottom — declining has to
// honour the route's own "to the bottom" instruction, which is the
// one piece of per-route bookkeeping that has to survive the pause.
func TestCommanderTuckedToBottomDeclineKeepsTheBottom(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	before := owner.Library.Size()
	if before == 0 {
		t.Fatal("test needs a non-empty library")
	}
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.TuckToLibraryForEffect(cmdID, true)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("TuckToLibraryForEffect: %v", err)
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, false); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Library, owner.Command, g.Battlefield)
	bottom, err := owner.Library.Bottom()
	if err != nil {
		t.Fatalf("Bottom: %v", err)
	}
	if bottom.InstanceID != cmdID {
		t.Error("to-the-bottom tuck did not survive the CR 903.9 pause")
	}
}

// --- mill -------------------------------------------------------

// TestCommanderMilledOffersCommandZone — a commander shuffled into a
// library and then milled is put into a graveyard, so 903.9 applies.
func TestCommanderMilledOffersCommandZone(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, owner.Library, owner)

	g.mu.Lock()
	err := g.MillNForEffect(owner.ID, 1)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("MillNForEffect: %v", err)
	}
	if owner.Graveyard.Contains(cmdID) {
		t.Fatalf("milled commander hit the graveyard before the prompt was answered")
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Graveyard, owner.Library)
}

// TestCommanderMilledMidStackDoesNotEatTheRestOfTheMill — the paused
// card sits in the library while the prompt is outstanding, so the
// mill loop must keep counting off the cards it decided on up front
// rather than tripping over the commander on top.
func TestCommanderMilledMidStackDoesNotEatTheRestOfTheMill(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	// Library top is the LAST element, so pushing the commander now
	// and two more after it leaves the commander third from the top.
	cmdID := seatCommander(t, owner.Library, owner)
	var after []uuid.UUID
	for i := 0; i < 2; i++ {
		id := uuid.New()
		owner.Library.PushTop(Card{
			InstanceID: id, Name: "Filler", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID,
		})
		after = append(after, id)
	}

	g.mu.Lock()
	err := g.MillNForEffect(owner.ID, 3)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("MillNForEffect: %v", err)
	}
	for _, id := range after {
		if !owner.Graveyard.Contains(id) {
			t.Errorf("non-commander card %s was not milled", id)
		}
	}
	prompt := expectCommanderPrompt(t, g, owner)
	if err := g.ResolveOptionalReplacement(prompt.ID, owner.ID, true); err != nil {
		t.Fatalf("ResolveOptionalReplacement: %v", err)
	}
	assertOnlyIn(t, cmdID, owner.Command, owner.Graveyard, owner.Library)
}

// hasEventFor reports whether the game log carries an event of `kind`
// naming `cardID`.
func hasEventFor(g *Game, kind EventKind, cardID uuid.UUID) bool {
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.CardID == cardID {
			return true
		}
	}
	return false
}

// --- #359: an Optional replacement must not pause an entry that has
// no resume ---------------------------------------------------------

// TestOptionalReplacementDoesNotStrandAnUnresumableEntry is #359.
//
// The CR 614.10 "may" branch of the apply-loop used to queue its
// yes/no prompt unconditionally, while the EntryLifeCost branch
// beside it has checked entryResumable since #268. A permanent
// entering the battlefield by a route that cannot be resumed — a
// reanimation, an exile-return, a library search — would therefore
// pause on an Optional self-replacement with nothing to finish the
// move: the card stays in its old zone and the prompt is answerable
// to no effect.
//
// The reanimation path is the vehicle here because it is the easiest
// non-resumable entry to drive; the defect is in the apply-loop, not
// in any one entry site.
func TestOptionalReplacementDoesNotStrandAnUnresumableEntry(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cardID := uuid.New()
	owner.Graveyard.PushTop(Card{
		InstanceID: cardID, Name: "Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID,
	})

	g.mu.Lock()
	g.RegisterReplacementForTest(ReplacementEffect{
		Watches:        []EventKind{EventZoneMove},
		Optional:       true,
		PromptQuestion: "Reveal a land from your hand?",
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == cardID &&
				ev.NewZone == ZoneBattlefield
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(_ *ReplacementEvent, _ *Game, _ *Card) uuid.UUID {
			return owner.ID
		},
		Label: "Test optional entry replacement",
	})
	err := g.ReturnFromGraveyardForEffect(cardID, ZoneBattlefield)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ReturnFromGraveyardForEffect: %v", err)
	}

	if len(g.PendingChoices) != 0 {
		t.Errorf("unresumable entry queued %d prompts; nothing can answer them", len(g.PendingChoices))
	}
	if !g.Battlefield.Contains(cardID) {
		t.Fatal("the permanent was stranded instead of entering the battlefield")
	}
	if owner.Graveyard.Contains(cardID) {
		t.Error("the permanent is in two places at once")
	}
	// The un-applied branch: weaker than printed, never stranded.
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == cardID && c.Tapped {
			t.Error("the un-answered replacement was applied anyway")
		}
	}
}

// TestOptionalReplacementStillPausesAnExit guards the narrowness of
// the #359 fix: the guard is about entries with no resume, and must
// not silence the CR 903.9 prompt on the exit routes, which do have
// one. Everything in this file depends on that, but this states it.
func TestOptionalReplacementStillPausesAnExit(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	cmdID := seatCommander(t, g.Battlefield, owner)

	g.mu.Lock()
	err := g.ExileCardForEffect(cmdID)
	g.mu.Unlock()
	if err != nil {
		t.Fatalf("ExileCardForEffect: %v", err)
	}
	expectCommanderPrompt(t, g, owner)
}
