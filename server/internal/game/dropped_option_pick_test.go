package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// dropped_option_pick_test.go — #1006, CR 800.4f and the #544 rule.
//
// An option pick is asked from inside a resolution that is PAUSED
// waiting for it, so its frame holds the rest of the card. Dropping
// the prompt used to throw that away: Fact or Fiction put neither pile
// anywhere, a Torment of Hailfire stopped at the victim who left, and
// a "choose a player" never ran the sentence printed after the choice.
//
// The fix is one row in the departure table (dropDefault) performed at
// one place (runChoiceDropActionLocked), so both ways a prompt can be
// dropped settle it the same way:
//
//   - the CHOOSER leaves the game, and the prompt cannot be reassigned
//     — the departure sweep, with CR 800.4f/g's gates in front of it;
//   - the engine WITHDRAWS the prompt because it has no legal answer
//     left — dropChoiceLocked, which is the door every prune in the
//     tree already goes through.

// optionPickWithAContinuation queues a two-branch option pick whose
// continuation records the index it was run with. `ran` counts the
// runs, `got` is the last index.
func optionPickWithAContinuation(t *testing.T, g *Game, p OptionPickPrompt, ran *int, got *int) uuid.UUID {
	t.Helper()
	inner := p.Then
	p.Then = func(g *Game, index int) error {
		*ran++
		*got = index
		if inner != nil {
			return inner(g, index)
		}
		return nil
	}
	if len(p.Options) == 0 {
		p.Options = []ChoiceOption{{Label: "Lose 3 life", LifeCost: 3}, {Label: "Discard a card"}}
	}
	var id uuid.UUID
	g.WithWriteLock(func() { id = g.QueueOptionPickForEffect(p) })
	if id == uuid.Nil {
		t.Fatalf("setup: the option pick was not queued")
	}
	return id
}

// TestADepartedChoosersOptionPickStillRunsTheRestOfTheCard is the
// bug's first cause, and Torment of Hailfire is the card: "each
// opponent loses 3 life unless that player sacrifices a nonland
// permanent or discards a card", repeated X times, asked strictly one
// opponent at a time so the next question is in the continuation.
//
// The prompt is about the departing opponent's OWN material, so
// CR 800.4g does not reassign it (TestOwnMaterialOptionPickIsDropped-
// NotReassigned, leave_game_choices_test.go, pins that). What used to
// go with the drop was every LATER opponent's question.
func TestADepartedChoosersOptionPickStillRunsTheRestOfTheCard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Torment of Hailfire")

	ran, got := 0, 0
	id := optionPickWithAContinuation(t, g, OptionPickPrompt{
		Chooser:  leaver.ID,
		Source:   source,
		Question: "Torment of Hailfire",
	}, &ran, &got)

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if c := findChoice(g, id); c != nil {
		t.Fatalf("an own-material option pick was reassigned to %s", c.Chooser)
	}
	if ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — the rest of the card was dropped with the question", ran)
	}
	if got != NoChoiceIndex {
		t.Errorf("the continuation ran with index %d, want NoChoiceIndex (%d): a drop must not "+
			"pick a branch on the departed chooser's behalf", got, NoChoiceIndex)
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, leaver.ID) == nil {
		t.Errorf("no EventPendingChoiceDropped for the dropped prompt")
	}
}

// The card whose text the continuation belongs to has to still be in
// the game. When it left with the chooser (CR 800.4a took both), the
// resolution is over and there is nothing to finish — the gate the
// departure sweep puts in front of the action
// (choiceObjectSurvivesLocked, leave_game.go).
func TestADepartedChoosersOwnCardsOptionPickRunsNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	source := departureTestSource(g, leaver.ID, "Fact or Fiction")

	ran, got := 0, 0
	optionPickWithAContinuation(t, g, OptionPickPrompt{
		Chooser:  leaver.ID,
		Source:   source,
		Question: "Fact or Fiction — take a pile",
	}, &ran, &got)

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if ran != 0 {
		t.Errorf("the continuation of a card that left the game with its controller ran %d times", ran)
	}
}

// The second cause: the engine WITHDRAWS the prompt while everybody
// concerned is still at the table. dropChoiceLocked is the one door
// every prune in the tree already goes through — a prompt whose option
// list has been emptied by a seat leaving it (CR 800.4a) reaches this
// exact line — so the rule is tested where every cause meets.
func TestAWithdrawnOptionPickStillRunsTheRestOfTheCard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Torment of Hailfire")

	ran, got := 0, 0
	id := optionPickWithAContinuation(t, g, OptionPickPrompt{
		Chooser:  chooser.ID,
		Source:   source,
		Question: "Torment of Hailfire",
	}, &ran, &got)

	g.WithWriteLock(func() {
		idx, _ := g.findChoiceLocked(id)
		if idx < 0 {
			t.Fatalf("setup: the prompt is not in the queue")
		}
		g.dropChoiceLocked(idx)
	})
	if findChoice(g, id) != nil {
		t.Fatal("the withdrawn prompt is still queued")
	}
	if ran != 1 || got != NoChoiceIndex {
		t.Fatalf("the continuation ran %d times with index %d, want once with NoChoiceIndex (%d)",
			ran, got, NoChoiceIndex)
	}
	// And the table is not left waiting on a question nobody can
	// answer, which is why the prompt was withdrawn in the first place.
	if err := g.PassPriority(); err != nil {
		t.Errorf("PassPriority after the prompt was withdrawn: %v", err)
	}
}

// An ANSWER is not a drop. The two exits from the queue share one line
// of slice surgery and nothing else — a dequeue that also ran the drop
// action would run the card's continuation twice, once for the branch
// the chooser picked and once for "nobody chose".
func TestAnsweringAnOptionPickRunsTheContinuationExactlyOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Torment of Hailfire")

	ran, got := 0, 0
	id := optionPickWithAContinuation(t, g, OptionPickPrompt{
		Chooser:  chooser.ID,
		Source:   source,
		Question: "Torment of Hailfire",
	}, &ran, &got)

	if err := g.ResolveOptionPick(id, chooser.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if ran != 1 || got != 1 {
		t.Errorf("the continuation ran %d times with index %d, want once with 1", ran, got)
	}
}

// A client cannot send the no-choice outcome. It is an engine-side
// settlement, not an answer, and ResolveOptionPick refuses it the way
// it refuses any other out-of-range index — with the prompt still
// open, so the seat can try again.
func TestAClientCannotAnswerWithTheNoChoiceIndex(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Torment of Hailfire")

	ran, got := 0, 0
	id := optionPickWithAContinuation(t, g, OptionPickPrompt{
		Chooser:  chooser.ID,
		Source:   source,
		Question: "Torment of Hailfire",
	}, &ran, &got)

	if err := g.ResolveOptionPick(id, chooser.ID, NoChoiceIndex); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("ResolveOptionPick(NoChoiceIndex) = %v, want ErrInvalidParam", err)
	}
	if ran != 0 {
		t.Errorf("a refused answer ran the continuation %d times", ran)
	}
	if findChoice(g, id) == nil {
		t.Error("a refused answer took the prompt with it")
	}
}

// The pile split, which is the shape #1006 was filed about. A dropped
// pick used to leave the five revealed cards where they were — neither
// pile into the hand, neither into the graveyard — because the
// continuation that moves them went with the prompt.
//
// The engine's own queue-time guards already take the first pile when
// the chooser cannot be asked (queuePilePickLocked); this is the same
// answer at the drop path, reached through the frame rather than by
// repeating the rule.
func TestADroppedPileSplitStillPutsThePilesSomewhere(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	owner, chooser := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, owner.ID, "Steam Augury")
	all := handCardOf(t, owner, 3)

	var taken, left []uuid.UUID
	ran := 0
	g.WithWriteLock(func() {
		if err := g.queuePilePickLocked(chooser.ID, owner.ID, source, "Take a pile",
			all, all[:1], func(_ *Game, tk, lf []uuid.UUID) error {
				ran++
				taken, left = tk, lf
				return nil
			}); err != nil {
			t.Fatalf("queuePilePickLocked: %v", err)
		}
	})
	c := onlyPendingChoice(t, g)
	if c.Kind != PendingChoiceOptionPick {
		t.Fatalf("prompt kind = %q, want option_pick", c.Kind)
	}

	g.WithWriteLock(func() { g.dropChoiceLocked(0) })
	if ran != 1 {
		t.Fatalf("the pile split's continuation ran %d times, want 1", ran)
	}
	if len(taken) != 1 || taken[0] != all[0] {
		t.Errorf("taken pile = %v, want the first pile %v", taken, all[:1])
	}
	if len(left) != len(all)-1 {
		t.Errorf("left pile = %v, want the other %d cards", left, len(all)-1)
	}
}

// And the player choice, whose continuation is "anything printed after
// the choice". A drop records the ABSENCE of an answer on the item —
// the same thing QueueChoosePlayerForEffect records when it cannot ask
// at all — so a card that asks twice does not read the first clause's
// player for the second.
func TestADroppedPlayerChoiceRecordsTheAbsenceAndRunsOn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Skullwinder")
	item := &StackItem{ID: uuid.New(), Controller: caster.ID, SourceCardID: source}

	ran := 0
	var chosen uuid.UUID
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser:  chooser.ID,
			Among:    []uuid.UUID{g.Seats[2].ID, g.Seats[3].ID},
			Source:   source,
			Question: "Choose an opponent",
			Item:     item,
			Then: func(_ *Game, seat uuid.UUID) error {
				ran++
				chosen = seat
				return nil
			},
		})
	})
	if id == uuid.Nil {
		t.Fatal("setup: the player choice was not queued")
	}

	g.WithWriteLock(func() {
		idx, _ := g.findChoiceLocked(id)
		g.dropChoiceLocked(idx)
	})
	if ran != 1 {
		t.Fatalf("the continuation after the player choice ran %d times, want 1", ran)
	}
	if chosen != uuid.Nil {
		t.Errorf("the continuation was handed %s; nobody was chosen", chosen)
	}
	if ChosenPlayerOn(item) != uuid.Nil {
		t.Errorf("the item records %s as chosen after the prompt was dropped", ChosenPlayerOn(item))
	}
	if len(item.Payload) != 1 || item.Payload[0].Kind != TargetNone {
		t.Errorf("payload = %+v, want one TargetNone — the absence has to be recorded, "+
			"or a second clause reads the first one's answer", item.Payload)
	}
}

// #994's seat prune is the OTHER way an option pick is dropped while
// everybody concerned is still at the table: a survivor's prompt whose
// every remaining option named a seat that has just left (CR 800.4a).
// It reaches dropChoiceLocked, so it picks up the default run with no
// wiring of its own — and this asserts that rather than assuming it,
// because the two landed a day apart.
//
// The continuation is a SEAT one (thenSeat, #994), so "nobody chose"
// is spelled uuid.Nil rather than as an index. Same absence, same one
// function (runWithNoChoice).
func TestAnOptionPickEmptiedByASeatPruneStillRunsTheRestOfTheCard(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, leaver := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Skullwinder")
	item := &StackItem{ID: uuid.New(), Controller: caster.ID, SourceCardID: source}

	ran := 0
	var chosen uuid.UUID
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChoosePlayerForEffect(ChoosePlayerPrompt{
			Chooser: caster.ID,
			// The ONLY candidate is the seat that is about to leave,
			// so the prune empties the question rather than shortening
			// it.
			Among:    []uuid.UUID{leaver.ID},
			Source:   source,
			Question: "Choose an opponent",
			Item:     item,
			Then: func(_ *Game, seat uuid.UUID) error {
				ran++
				chosen = seat
				return nil
			},
		})
	})
	if id == uuid.Nil {
		t.Fatal("setup: the player choice was not queued")
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if findChoice(g, id) != nil {
		t.Fatal("a prompt whose every option named the departed seat is still queued")
	}
	if ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1 — the chooser is still at the table "+
			"and the rest of their card still has to resolve", ran)
	}
	if chosen != uuid.Nil {
		t.Errorf("the continuation was handed %s; nobody was left to choose", chosen)
	}
	if ChosenPlayerOn(item) != uuid.Nil {
		t.Errorf("the item records %s as chosen after the prompt was emptied", ChosenPlayerOn(item))
	}
	// The prompt blocked the table while it was queued, so the drop has
	// to free it — that is the half this whole sprint is about.
	if err := g.PassPriority(); err != nil {
		t.Errorf("PassPriority after the emptied prompt was dropped: %v", err)
	}
}

// uuid.Nil means "nobody chose" only because an ANSWER can never carry
// it: ResolveOptionPick refuses an option that names no seat before the
// frame runs, so a seat continuation cannot be handed the absence by a
// client. Without this the two meanings would collide on one value.
func TestASeatContinuationRefusesAnOptionThatNamesNoSeat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	caster, chooser := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, caster.ID, "Skullwinder")

	ran := 0
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueOptionPickForEffect(OptionPickPrompt{
			Chooser:  chooser.ID,
			Source:   source,
			Question: "Choose an opponent",
			Options: []ChoiceOption{
				{Label: "Seat with an id", Player: g.Seats[2].ID},
				{Label: "Seat with none"},
			},
			ThenSeat: func(_ *Game, _ uuid.UUID) error { ran++; return nil },
		})
	})
	if id == uuid.Nil {
		t.Fatal("setup: the prompt was not queued")
	}
	if err := g.ResolveOptionPick(id, chooser.ID, 1); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("answering a seat continuation with a seatless option = %v, want ErrInvalidParam", err)
	}
	if ran != 0 {
		t.Errorf("the seat continuation ran %d times on a refused answer", ran)
	}
	if findChoice(g, id) == nil {
		t.Error("a refused answer took the prompt with it; the chooser can no longer try again")
	}
	// The good option still answers.
	if err := g.ResolveOptionPick(id, chooser.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick on the seat option: %v", err)
	}
	if ran != 1 {
		t.Errorf("the seat continuation ran %d times on a good answer, want 1", ran)
	}
}
