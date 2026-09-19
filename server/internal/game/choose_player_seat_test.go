package game

import (
	"testing"

	"github.com/google/uuid"
)

// choose_player_seat_test.go — #994. An option can name the SEAT it is
// about, and a seat that has left the game comes off an open prompt.
//
// CR 800.4a: a player who has left the game is not a player, so
// "choose a player" cannot name them. Eligibility used to be filtered
// exactly once, at queue time (eligibleChosenPlayersLocked); these are
// the two moments after that — the chooser leaving, and one of the
// OFFERED players leaving.

// optionSeats is the seat each option is about, in offer order.
func optionSeats(c *PendingChoice) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(c.PickOptions))
	for _, opt := range c.PickOptions {
		out = append(out, opt.Player)
	}
	return out
}

// TestASeatOptionCarriesTheSeatItLabels is the field the rest of #994
// rests on: the engine can tell WHICH player an option is about
// without string-matching the rendered name.
func TestASeatOptionCarriesTheSeatItLabels(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser:  me.ID,
			Among:    []uuid.UUID{g.Seats[1].ID, g.Seats[2].ID},
			Question: "test — choose a player",
		})
	})
	c := choosePlayerPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	for _, opt := range c.PickOptions {
		if opt.Player == uuid.Nil {
			t.Fatalf("a seat option carries its seat: %+v", opt)
		}
		if got := g.PlayerByID(opt.Player); got == nil || got.Name != opt.Label {
			t.Errorf("option %q is about %s; the label and the seat must agree", opt.Label, opt.Player)
		}
	}
}

// TestAnOptionThatIsNotAboutASeatLeavesTheFieldZero — every option the
// engine built before #994 keeps a zero Player, which is what makes the
// pruning paths safe for them: keepSeats never asks about an option
// that names no seat.
func TestAnOptionThatIsNotAboutASeatLeavesTheFieldZero(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:  me.ID,
			Question: "Torment of Hailfire",
			Options: []ChoiceOption{
				{Label: "Lose 3 life", LifeCost: 3},
				{Label: "Discard a card"},
			},
			Then: func(*Game, int) error { return nil },
		})
	})
	c := choosePlayerPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	for _, opt := range c.PickOptions {
		if opt.Player != uuid.Nil {
			t.Errorf("a consequence option is about no seat: %+v", opt)
		}
	}
}

// TestAnOfferedSeatThatLeavesIsPrunedFromAnOpenPrompt is #994's repro,
// and the bug it reports: four seats, a prompt open at seat 0 offering
// three of them, and one of the OFFERED players leaves before it is
// answered. The prompt kept offering them.
//
// The chooser is not the one who leaves — that is the other moment, and
// the reassignment handles it (below).
func TestAnOfferedSeatThatLeavesIsPrunedFromAnOpenPrompt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, leaver := g.Seats[0], g.Seats[2]
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser:  me.ID,
			Among:    []uuid.UUID{g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID},
			Question: "test — choose a player",
		})
	})
	c := choosePlayerPromptFor(g, me.ID)
	if c == nil || len(c.PickOptions) != 3 {
		t.Fatalf("the prompt starts with three seats: %+v", c)
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	c = choosePlayerPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("the prompt is still owed by a seat that is still here: %+v", g.PendingChoices)
	}
	for _, seat := range optionSeats(c) {
		if seat == leaver.ID {
			t.Fatalf("a departed seat is still offered (CR 800.4a): %v", optionLabels(c))
		}
	}
	if len(c.PickOptions) != 2 {
		t.Fatalf("the other two seats are untouched: %v", optionLabels(c))
	}
}

// TestPruningASeatDoesNotRenumberTheAnswer is the half a prune could
// have broken on its own: the chooser picks the SECOND remaining
// option, and the player recorded must be that option's seat and not
// whoever was second on the list the prompt was built with.
func TestPruningASeatDoesNotRenumberTheAnswer(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, leaver := g.Seats[0], g.Seats[1]
	// Most life first, so the offer order is seats 1, 2, 3.
	g.Seats[1].Life, g.Seats[2].Life, g.Seats[3].Life = 40, 30, 20

	item := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, Controller: me.ID}
	var got uuid.UUID
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser:  me.ID,
			Among:    []uuid.UUID{g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID},
			Question: "test — choose a player",
			Item:     item,
			Then:     func(_ *Game, chosen uuid.UUID) error { got = chosen; return nil },
		})
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := choosePlayerPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	// Index 1 of what is LEFT is seat 3. Before the prune it was seat 2.
	if err := g.ResolveOptionPick(c.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if got != g.Seats[3].ID {
		t.Errorf("chose %s, want seat 3 (%s) — the answer is the option's own seat, not an index into the list it was built with",
			got, g.Seats[3].ID)
	}
	if rec := ChosenPlayerOn(item); rec != g.Seats[3].ID {
		t.Errorf("recorded %s on the item, want seat 3 (%s)", rec, g.Seats[3].ID)
	}
}

// TestAChooserWhoLeavesTakesTheirOwnSeatOffTheReassignedPrompt — the
// other moment. CR 800.4g hands the prompt to somebody else, and the
// list it hands over must not still offer the seat that just left.
func TestAChooserWhoLeavesTakesTheirOwnSeatOffTheReassignedPrompt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver, owner := g.Seats[1], g.Seats[0]
	// The prompt must be about somebody else's object for CR 800.4g to
	// reassign it at all (departedChoiceObjectLocked).
	source := departureTestSource(g, owner.ID, "Reassignment Source")
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:    leaver.ID,
			FromPlayer: owner.ID,
			Source:     source,
			Question:   "test — choose a player",
			Options:    g.seatChoiceOptionsLocked([]uuid.UUID{leaver.ID, g.Seats[2].ID, g.Seats[3].ID}),
			ThenSeat:   func(*Game, uuid.UUID) error { return nil },
		})
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	var c *PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == PendingChoiceOptionPick {
			c = pc
		}
	}
	if c == nil {
		t.Fatalf("the prompt was reassigned, not dropped: %+v", g.PendingChoices)
	}
	if c.Chooser == leaver.ID {
		t.Fatalf("it still belongs to the seat that left: %s", c.Chooser)
	}
	for _, seat := range optionSeats(c) {
		if seat == leaver.ID {
			t.Fatalf("the reassigned prompt still offers the departed seat: %v", optionLabels(c))
		}
	}
	if len(c.PickOptions) != 2 {
		t.Errorf("the surviving seats are untouched: %v", optionLabels(c))
	}
}

// TestAPromptWhoseEverySeatHasLeftIsDropped — pruning to nothing is the
// state QueueChoosePlayerForEffect refuses to queue in the first place,
// and an option_pick blocks the table, so an empty one is the #544
// wedge rather than a harmless leftover.
func TestAPromptWhoseEverySeatHasLeftIsDropped(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, leaver := g.Seats[0], g.Seats[2]
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser:  me.ID,
			Among:    []uuid.UUID{leaver.ID},
			Question: "test — choose a player",
		})
	})
	if choosePlayerPromptFor(g, me.ID) == nil {
		t.Fatalf("the prompt was queued: %+v", g.PendingChoices)
	}
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if c := choosePlayerPromptFor(g, me.ID); c != nil {
		t.Fatalf("an unanswerable prompt was left in the queue: %+v", c.PickOptions)
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, me.ID) == nil {
		t.Error("the drop is announced, so a stall dump says where the prompt went")
	}
}

// TestTheSeatPruneLeavesAPileSplitAlone — keepSeats is asked only about
// options that name a seat, so Fact or Fiction's two piles survive a
// departure they have nothing to do with. An option list with a
// non-seat branch in it can never be emptied by the prune.
func TestTheSeatPruneLeavesAPileSplitAlone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, leaver := g.Seats[0], g.Seats[2]
	pile := handCardOf(t, g.Seats[0], 2)
	g.WithWriteLock(func() {
		g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:  me.ID,
			Question: "Fact or Fiction",
			Options: []ChoiceOption{
				{Label: "Take pile 1", Cards: pile[:1]},
				{Label: "Take pile 2", Cards: pile[1:]},
			},
			Then: func(*Game, int) error { return nil },
		})
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := choosePlayerPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("the pile split is untouched by a departure: %+v", g.PendingChoices)
	}
	if len(c.PickOptions) != 2 {
		t.Errorf("both piles are still offered: %v", optionLabels(c))
	}
}
