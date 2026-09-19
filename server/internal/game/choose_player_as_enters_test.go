package game

import (
	"testing"

	"github.com/google/uuid"
)

// choose_player_as_enters_test.go — #980, CR 614.12. The THIRD form of
// "choose a player": made as the permanent enters, stored on it, and
// read for the rest of its life.
//
// What it shares with #929's resolution-time form is the PROMPT, and
// these tests are mostly about that sharing: the same option_pick kind,
// the same seat options, and therefore the same CR 800.4a pruning. What
// it does NOT share is where the answer goes, which is the first test
// below and the reason the form exists.

// pushNemesisPermanent puts a bare permanent on the battlefield for the
// as-enters prompt to be about. The catalog hook is not involved — this
// is the engine half, and the card has its own tests.
func pushNemesisPermanent(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Merfolk Rogue"
	c.Power, c.Toughness = 3, 1
	c.Keywords = []string{ProtectionFromChosenPlayer}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// TestTheAsEntersPromptRidesTheExistingKind — no new PendingChoiceKind,
// which is what makes every consumer answer it unchanged: the choice
// gate, the enumerator, the wire projection and the client modal.
func TestTheAsEntersPromptRidesTheExistingKind(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[1]
	source := pushNemesisPermanent(g, owner, "True-Name Nemesis")

	g.WithWriteLock(func() {
		g.QueueChoosePlayerAsEntersForEffect(owner.ID, source, "True-Name Nemesis — choose a player",
			[]uuid.UUID{g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID})
	})
	c := choosePlayerPromptFor(g, owner.ID)
	if c == nil {
		t.Fatalf("the prompt is addressed to the entering permanent's controller: %+v", g.PendingChoices)
	}
	if c.Kind != PendingChoiceOptionPick {
		t.Errorf("kind %s, want the existing option_pick", c.Kind)
	}
	if !ChoiceBlocksTable(c.Kind) {
		t.Error("an unanswered as-enters choice stops the table")
	}
	if c.Source != source {
		t.Errorf("the prompt names the entering permanent: %v", c.Source)
	}
	// "Choose a player" is every seat, the controller included (CR
	// 614.12 names no restriction). Offering yourself is legal and
	// useless, and the engine does not invent a narrowing.
	if len(c.PickOptions) != 4 {
		t.Fatalf("every seat is offered: %v", optionLabels(c))
	}
	var sawController bool
	for _, seat := range optionSeats(c) {
		if seat == owner.ID {
			sawController = true
		}
	}
	if !sawController {
		t.Error("the controller is a legal answer")
	}
}

// TestAnsweringTheAsEntersPromptStampsThePermanent is the difference
// from #929's form: the answer does not live on a stack item for one
// resolution, it lives on the permanent.
func TestAnsweringTheAsEntersPromptStampsThePermanent(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, victim := g.Seats[1], g.Seats[2]
	source := pushNemesisPermanent(g, owner, "True-Name Nemesis")

	g.WithWriteLock(func() {
		g.QueueChoosePlayerAsEntersForEffect(owner.ID, source, "choose a player",
			[]uuid.UUID{victim.ID, g.Seats[3].ID})
	})
	c := choosePlayerPromptFor(g, owner.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	idx := -1
	for i, opt := range c.PickOptions {
		if opt.Player == victim.ID {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("the victim is offered: %v", optionLabels(c))
	}
	if err := g.ResolveOptionPick(c.ID, owner.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if got := g.ChosenPlayerOf(source); got != victim.ID {
		t.Errorf("stored %v, want %v", got, victim.ID)
	}
	// And it is announced, the way a chosen colour and a named tribe
	// are: the prompt closes and the log is the only record left.
	var logged *Event
	for i := len(g.Events) - 1; i >= 0; i-- {
		if g.Events[i].Kind == EventPlayerChosen {
			logged = &g.Events[i]
			break
		}
	}
	if logged == nil {
		t.Fatal("the answer is recorded in the event log")
	}
	if logged.CardID != source || logged.Target != victim.ID {
		t.Errorf("the log entry says which permanent chose whom: %+v", logged)
	}
	if logged.Label != victim.Name {
		t.Errorf("label %q, want the seat's name %q", logged.Label, victim.Name)
	}
}

// TestADepartedSeatIsPrunedFromTheAsEntersPromptToo — the point of
// building this on the same option shape (#994). A Nemesis asking
// while somebody concedes must stop offering them, and it gets that
// for free rather than through a second copy of the rule.
func TestADepartedSeatIsPrunedFromTheAsEntersPromptToo(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, leaver := g.Seats[1], g.Seats[2]
	source := pushNemesisPermanent(g, owner, "True-Name Nemesis")

	g.WithWriteLock(func() {
		g.QueueChoosePlayerAsEntersForEffect(owner.ID, source, "choose a player",
			[]uuid.UUID{g.Seats[0].ID, leaver.ID, g.Seats[3].ID})
	})
	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	c := choosePlayerPromptFor(g, owner.ID)
	if c == nil {
		t.Fatalf("the prompt is still owed: %+v", g.PendingChoices)
	}
	for _, seat := range optionSeats(c) {
		if seat == leaver.ID {
			t.Fatalf("a departed seat is still offered (CR 800.4a): %v", optionLabels(c))
		}
	}
	if len(c.PickOptions) != 2 {
		t.Errorf("the other two seats are untouched: %v", optionLabels(c))
	}
}

// TestTheAsEntersAnswerLandsNowhereWhenThePermanentHasGone — the prompt
// is asynchronous and the Nemesis can be killed in response to its own
// entry. The choice is still made and simply has nowhere to go, which
// is what CR 608.2 does with an effect whose object has left; it must
// not be an error and must not wedge the seat.
func TestTheAsEntersAnswerLandsNowhereWhenThePermanentHasGone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, victim := g.Seats[1], g.Seats[2]
	source := pushNemesisPermanent(g, owner, "True-Name Nemesis")

	g.WithWriteLock(func() {
		g.QueueChoosePlayerAsEntersForEffect(owner.ID, source, "choose a player", []uuid.UUID{victim.ID})
	})
	c := choosePlayerPromptFor(g, owner.ID)
	if c == nil {
		t.Fatalf("no prompt: %+v", g.PendingChoices)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(source); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if err := g.ResolveOptionPick(c.ID, owner.ID, 0); err != nil {
		t.Fatalf("answering a prompt whose permanent has gone is not an error: %v", err)
	}
	if choosePlayerPromptFor(g, owner.ID) != nil {
		t.Error("the prompt is gone either way")
	}
}

// TestTheAsEntersPromptIsNotQueuedWithNobodyToChoose — the whole table
// has left, so there is no question. ChosenPlayer stays nil, which
// reads as "protected from nobody".
func TestTheAsEntersPromptIsNotQueuedWithNobodyToChoose(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner := g.Seats[1]
	source := pushNemesisPermanent(g, owner, "True-Name Nemesis")

	var queued uuid.UUID
	g.WithWriteLock(func() {
		queued = g.QueueChoosePlayerAsEntersForEffect(owner.ID, source, "choose a player", nil)
	})
	if queued != uuid.Nil {
		t.Errorf("a question with no answers is not asked: %v", queued)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("nothing was queued: %+v", g.PendingChoices)
	}
	if got := g.ChosenPlayerOf(source); got != uuid.Nil {
		t.Errorf("nobody was chosen: %v", got)
	}
}
