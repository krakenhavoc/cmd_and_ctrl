package game

import (
	"testing"

	"github.com/google/uuid"
)

// choose_player_test.go — #929. The resolution-time player choice,
// tested as a SHAPE rather than only through the cards that use it:
// which seats are offered, in what order, where the answer is
// recorded, what happens when nobody can be asked, and that the whole
// thing rewinds.

// choosePlayerPromptFor returns the newest open option pick owed by a
// seat, or nil.
func choosePlayerPromptFor(g *Game, chooser uuid.UUID) *PendingChoice {
	var out *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceOptionPick && c.Chooser == chooser {
			out = c
		}
	}
	return out
}

func optionLabels(c *PendingChoice) []string {
	out := make([]string, 0, len(c.PickOptions))
	for _, opt := range c.PickOptions {
		out = append(out, opt.Label)
	}
	return out
}

// TestChoosePlayerOffersExactlyTheEligibleSeats — the candidate list
// is narrowed to seats that can actually be chosen, the labels are
// the seat names, and the order is the documented bot policy (most
// life first).
func TestChoosePlayerOffersExactlyTheEligibleSeats(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	// A seat that has left the game is not a player (CR 800.4a) and
	// cannot be chosen.
	g.Seats[3].Eliminated = true
	g.Seats[1].Life = 12
	g.Seats[2].Life = 33

	item := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, Controller: me.ID}
	var got uuid.UUID
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser:  me.ID,
			Among:    []uuid.UUID{g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID, g.Seats[1].ID},
			Question: "test — choose a player",
			Item:     item,
			Then:     func(_ *Game, chosen uuid.UUID) error { got = chosen; return nil },
		})
	})

	c := choosePlayerPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("the prompt is addressed to the chooser: %+v", g.PendingChoices)
	}
	if c.Kind != PendingChoiceOptionPick {
		t.Errorf("it rides the existing kind: %s", c.Kind)
	}
	want := []string{g.Seats[2].Name, g.Seats[1].Name}
	labels := optionLabels(c)
	if len(labels) != len(want) {
		t.Fatalf("offered %v, want exactly the two eligible seats %v", labels, want)
	}
	for i := range want {
		if labels[i] != want[i] {
			t.Errorf("option %d = %q, want %q (most life first)", i, labels[i], want[i])
		}
	}
	for _, opt := range c.PickOptions {
		if len(opt.Cards) != 0 || opt.LifeCost != 0 {
			t.Errorf("a seat option names no cards and costs no life: %+v", opt)
		}
	}

	if err := g.ResolveOptionPick(c.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if got != g.Seats[1].ID {
		t.Errorf("Then got %s, want the seat at index 1 (%s)", got, g.Seats[1].ID)
	}
	if ChosenPlayerOn(item) != g.Seats[1].ID {
		t.Errorf("the answer is recorded on the item: %v", item.Payload)
	}
	if len(ChosenPlayersOn(item)) != 1 {
		t.Errorf("one question, one recorded answer: %v", item.Payload)
	}
}

// TestChoosePlayerRecordsEachAnswerInOrder — "choose a second player"
// reads the first answer back, and the payload keeps both in the
// order the card asked.
func TestChoosePlayerRecordsEachAnswerInOrder(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	item := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, Controller: me.ID}
	ask := func(among []uuid.UUID) {
		g.WithWriteLock(func() {
			g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
				Chooser: me.ID, Among: among, Question: "test", Item: item,
			})
		})
		c := choosePlayerPromptFor(g, me.ID)
		if c == nil {
			t.Fatalf("no prompt: %+v", g.PendingChoices)
		}
		if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
			t.Fatalf("ResolveOptionPick: %v", err)
		}
	}
	all := []uuid.UUID{g.Seats[0].ID, g.Seats[1].ID, g.Seats[2].ID, g.Seats[3].ID}
	g.Seats[0].Life = 50
	ask(all)
	first := ChosenPlayerOn(item)
	if first != g.Seats[0].ID {
		t.Fatalf("first answer %s, want the highest-life seat %s", first, g.Seats[0].ID)
	}
	// "A second player" — the caller excludes what is already chosen.
	var rest []uuid.UUID
	for _, id := range all {
		if id != first {
			rest = append(rest, id)
		}
	}
	ask(rest)
	chosen := ChosenPlayersOn(item)
	if len(chosen) != 2 || chosen[0] != first || chosen[1] == first {
		t.Errorf("payload %v, want two distinct answers in ask order", chosen)
	}
	if ChosenPlayerOn(item) != chosen[1] {
		t.Error("ChosenPlayerOn is the most recent answer")
	}
}

// TestChoosePlayerWithNobodyToAskRecordsTheAbsence — an empty pool
// queues nothing, and the item says so rather than keeping the
// previous clause's answer. Without the marker a card that asks twice
// would silently apply its second clause to its first clause's player.
func TestChoosePlayerWithNobodyToAskRecordsTheAbsence(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	item := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, Controller: me.ID}
	var queued uuid.UUID
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser: me.ID, Among: []uuid.UUID{g.Seats[1].ID}, Question: "test", Item: item,
		})
	})
	c := choosePlayerPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("the first question is asked")
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if ChosenPlayerOn(item) != g.Seats[1].ID {
		t.Fatalf("first answer not recorded: %v", item.Payload)
	}

	g.WithWriteLock(func() {
		queued = g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser: me.ID, Among: nil, Question: "test — nobody left", Item: item,
		})
	})
	if queued != uuid.Nil {
		t.Errorf("an empty pool queues nothing, got %s", queued)
	}
	if got := ChosenPlayerOn(item); got != uuid.Nil {
		t.Errorf("ChosenPlayerOn = %s, want nil — the absence must not read as the earlier answer", got)
	}
	if len(ChosenPlayersOn(item)) != 1 {
		t.Errorf("an absence excludes nobody from a later clause: %v", ChosenPlayersOn(item))
	}
}

// TestChoosePlayerDropsWhenTheChooserHasLeft — a chooser who is gone
// gets no prompt (CR 800.4a), the queue stays clean, and the absence
// is recorded. Reassigning the choice to another seat is #902's
// single reassignment function, not this prompt's business.
func TestChoosePlayerDropsWhenTheChooserHasLeft(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	gone, other := g.Seats[0], g.Seats[1]
	gone.Eliminated = true
	item := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, Controller: gone.ID}
	ran := false
	var queued uuid.UUID
	g.WithWriteLock(func() {
		queued = g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser:  gone.ID,
			Among:    []uuid.UUID{other.ID},
			Question: "test — a departed chooser",
			Item:     item,
			Then:     func(*Game, uuid.UUID) error { ran = true; return nil },
		})
	})
	if queued != uuid.Nil {
		t.Errorf("queued %s, want nothing for a seat that has left", queued)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("no prompt is left behind: %+v", g.PendingChoices)
	}
	if ran {
		t.Error("Then does not run from the engine side when nothing was asked")
	}
	if ChosenPlayerOn(item) != uuid.Nil {
		t.Errorf("the absence is recorded: %v", item.Payload)
	}
}

// TestChoosePlayerUndoesAcrossThePrompt — rewinding past the answer
// puts the question back with the same seats offered.
func TestChoosePlayerUndoesAcrossThePrompt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	item := &StackItem{ID: uuid.New(), Kind: StackItemTriggered, Controller: me.ID}
	g.WithWriteLock(func() {
		g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser:  me.ID,
			Among:    []uuid.UUID{g.Seats[1].ID, g.Seats[2].ID},
			Question: "test — choose a player",
			Item:     item,
			Then: func(g *Game, chosen uuid.UUID) error {
				return g.ChangePlayerLifeForEffect(uuid.Nil, chosen, -5)
			},
		})
	})
	before := g.Seats[1].Life
	snap := g.Clone()
	c := choosePlayerPromptFor(g, me.ID)
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if g.Seats[1].Life != before-5 {
		t.Fatalf("the branch ran: life %d, want %d", g.Seats[1].Life, before-5)
	}
	g.RestoreFrom(snap)
	if g.Seats[1].Life != before {
		t.Errorf("undo rewinds the branch: life %d, want %d", g.Seats[1].Life, before)
	}
	back := choosePlayerPromptFor(g, me.ID)
	if back == nil {
		t.Fatalf("undo puts the question back: %+v", g.PendingChoices)
	}
	if len(back.PickOptions) != 2 {
		t.Errorf("the restored prompt still offers both seats: %+v", back.PickOptions)
	}
	if err := g.ResolveOptionPick(back.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick after undo: %v", err)
	}
	if g.Seats[1].Life != before-5 {
		t.Errorf("answering the restored prompt: life %d, want %d", g.Seats[1].Life, before-5)
	}
}

// TestChoosePlayerLabelsANamelessSeat — an option button with no text
// is unanswerable, so a seat with no name still gets one.
func TestChoosePlayerLabelsANamelessSeat(t *testing.T) {
	g := newActiveGame(t)
	g.Seats[1].Name = ""
	var label string
	g.WithWriteLock(func() { label = g.seatLabelLocked(g.Seats[1].ID) })
	if label != "Seat 2" {
		t.Errorf("label %q, want %q", label, "Seat 2")
	}
	var missing string
	g.WithWriteLock(func() { missing = g.seatLabelLocked(uuid.New()) })
	if missing == "" {
		t.Error("even an unknown seat gets a label")
	}
}
