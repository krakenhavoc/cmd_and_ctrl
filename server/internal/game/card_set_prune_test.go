package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// card_set_prune_test.go pins #1045: a choose-cards prompt whose
// candidates leave the zone it picks from is TRIMMED, and one whose
// candidates have all gone is WITHDRAWN — settling its run leg on the
// way out — rather than left as a question with no legal answer.
//
// The hole this closes. A discard prompt is a PendingChoiceChooseCards
// over a FROZEN candidate list (the hand as it stood when the effect
// asked), and checkChooseCardsPicksLocked re-checks every pick against
// the LIVE zone before it accepts an answer. If every candidate leaves
// the hand while the prompt is open and its floor is one, no answer is
// legal: the prompt blocks advance_step / pass_priority / pass_turn
// (#791) and the bot enumerator has no set to offer, so the table is
// held (#544) — and since #1027 the run leg never settles either, so
// the rest of the printed instruction never runs (Syphon Mind's draw,
// Archon of Cruelty's last three clauses).
//
// The prune is ONE function keyed by the frame's zone, so the three
// hand shapes below and the graveyard and library picks are the same
// code path with a different ZoneKind.

// discardRunOverOneSeat starts a one-seat discard run and records what
// its continuation is handed.
func discardRunOverOneSeat(t *testing.T, g *Game, p DiscardPrompt) *discardRunCall {
	t.Helper()
	out := &discardRunCall{}
	g.WithWriteLock(func() {
		err := g.PlayerDiscardsThenForEffect(p, func(_ *Game, discarded PromptedDiscards) error {
			out.ran++
			out.got = discarded
			return nil
		})
		if err != nil {
			t.Fatalf("PlayerDiscardsThenForEffect: %v", err)
		}
	})
	return out
}

// wheelAwayTheHand is "each player discards their hand" arriving at one
// seat: the whole hand, at random, through the real discard path.
func wheelAwayTheHand(t *testing.T, g *Game, p *Player) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(p.ID, p.Hand.Size()); err != nil {
			t.Fatalf("DiscardRandomForEffect: %v", err)
		}
	})
}

// assertTableIsFree is the #544 assertion: the prompt is not merely
// gone from the queue, the seats can act again.
func assertTableIsFree(t *testing.T, g *Game) {
	t.Helper()
	if err := g.PassPriority(); errors.Is(err, ErrChoicePending) {
		t.Errorf("the table is still blocked behind the withdrawn prompt: %v", err)
	}
}

// --- the three hand shapes -------------------------------------------

// TestAWheeledHandWithdrawsItsDiscardPromptAndSettlesItsLeg is the
// issue's first shape: a Wheel resolves while a discard prompt is open
// and every candidate goes to the graveyard underneath it.
func TestAWheeledHandWithdrawsItsDiscardPromptAndSettlesItsLeg(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Archon of Cruelty")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	first := handCardForDiscard(t, g, me, "First")
	second := handCardForDiscard(t, g, me, "Second")

	run := discardRunOverOneSeat(t, g, DiscardPrompt{Player: me.ID, Source: source, N: 1})
	if c := discardPromptFor(g, me.ID); c == nil || len(c.ChooseCards) != 2 {
		t.Fatalf("setup: expected a prompt over two candidates, got %v", c)
	}

	wheelAwayTheHand(t, g, me)

	if c := discardPromptFor(g, me.ID); c != nil {
		t.Fatalf("the prompt survives with %d candidates and none of them in the hand — "+
			"no answer it would accept exists", len(c.ChooseCards))
	}
	if lastChoiceEvent(g, EventPendingChoiceDropped, me.ID) == nil {
		t.Error("the withdrawal is not in the log; a stall dump has to say where the prompt went")
	}
	if run.ran != 1 {
		t.Fatalf("the run's continuation ran %d times, want 1 — a withdrawn leg settles with "+
			"nothing discarded, it does not strand the instruction", run.ran)
	}
	if run.got.Discarded(me.ID) {
		t.Errorf("the seat is reported as having discarded through the PROMPT: %v", run.got)
	}
	if len(run.got) != 1 {
		t.Errorf("the seat was ASKED, so it has an entry: %v", run.got)
	}
	if !me.Graveyard.Contains(first) || !me.Graveyard.Contains(second) {
		t.Error("the wheel's own discards still happened")
	}
	assertTableIsFree(t, g)
}

// TestAMadnessExileUnderAnOpenDiscardPromptWithdrawsIt is the second
// shape, and the one that proves the prune asks about the ZONE rather
// than about the graveyard: CR 702.35a sends the card to exile, and
// what matters to the prompt is only that the hand no longer holds it.
func TestAMadnessExileUnderAnOpenDiscardPromptWithdrawsIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	withMadnessCard(t, "{R}")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := seedMadnessCard(me, "Fiery Temper", "Instant", "{1}{R}{R}")

	run := discardRunOverOneSeat(t, g, DiscardPrompt{Player: me.ID, Source: source, N: 1})
	if discardPromptFor(g, me.ID) == nil {
		t.Fatal("setup: no prompt over the one-card hand")
	}

	// Somebody else's effect pitches the only candidate first.
	discardOne(t, g, me, card, DiscardCauseEffect)

	if exiledCardByIDLocked(g, card) == nil {
		t.Fatal("setup: the madness card is not in exile")
	}
	if c := discardPromptFor(g, me.ID); c != nil {
		t.Error("the prompt survives a hand whose only card is in exile")
	}
	if run.ran != 1 {
		t.Fatalf("the run's continuation ran %d times, want 1", run.ran)
	}
	if run.got.Discarded(me.ID) {
		t.Errorf("nothing was discarded THROUGH THE PROMPT: %v", run.got)
	}
	assertTableIsFree(t, g)
}

// TestAHandBouncedOutFromUnderADiscardPromptWithdrawsIt is the third
// shape: the hand goes to the library, which is neither a discard nor
// a graveyard arrival — the prune keys on the candidate's zone and not
// on the kind of move that took it.
func TestAHandBouncedOutFromUnderADiscardPromptWithdrawsIt(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := handCardForDiscard(t, g, me, "Bounced")

	run := discardRunOverOneSeat(t, g, DiscardPrompt{Player: me.ID, Source: source, N: 1})
	if discardPromptFor(g, me.ID) == nil {
		t.Fatal("setup: no prompt over the one-card hand")
	}

	if err := g.MoveCardByID(
		ZoneRef{Kind: ZoneHand, Owner: me.ID},
		ZoneRef{Kind: ZoneLibrary, Owner: me.ID},
		card,
	); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}

	if c := discardPromptFor(g, me.ID); c != nil {
		t.Error("the prompt survives a hand whose only card is in the library")
	}
	if run.ran != 1 {
		t.Fatalf("the run's continuation ran %d times, want 1", run.ran)
	}
	if run.got.Discarded(me.ID) {
		t.Errorf("a card put into a library was not discarded: %v", run.got)
	}
	assertTableIsFree(t, g)
}

// --- the run, and the trim -------------------------------------------

// TestARunWhoseLegWasWithdrawnStillCompletes is the #1027 half taken to
// the end: one seat's hand is emptied under its prompt, the other seat
// answers, and the instruction finishes ONCE with what really left the
// hands. Before the prune the withdrawn seat's leg never settled, so
// the run's continuation — Syphon Mind's draw — never ran at all.
func TestARunWhoseLegWasWithdrawnStillCompletes(t *testing.T) {
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	source := departureTestSource(g, me.ID, "Syphon Mind")
	mine := oneCardHand(t, g, me, "Mine")
	oneCardHand(t, g, them, "Theirs")

	run := eachPlayerDiscardsThen(t, g, source, uuid.Nil, 1)
	if openDiscardPrompts(g) != 2 {
		t.Fatalf("setup: expected two prompts, got %d", openDiscardPrompts(g))
	}

	wheelAwayTheHand(t, g, them)

	if run.ran != 0 {
		t.Fatalf("the continuation ran %d times with the other seat's prompt still open, want 0",
			run.ran)
	}
	if openDiscardPrompts(g) != 1 {
		t.Fatalf("the emptied seat's prompt is still queued (%d open)", openDiscardPrompts(g))
	}

	answerDiscard(t, g, me.ID, mine)

	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want exactly 1", run.ran)
	}
	if run.got.Count() != 1 {
		t.Errorf("the run discarded %d cards, want 1 — only the seat that answered pitched (%v)",
			run.got.Count(), run.got)
	}
	if !run.got.Discarded(me.ID) {
		t.Error("the seat that answered discarded its card")
	}
	if run.got.Discarded(them.ID) {
		t.Errorf("the withdrawn seat is reported as having discarded something: %v", run.got)
	}
	if len(run.got) != 2 {
		t.Errorf("both seats were ASKED, so both have an entry: %v", run.got)
	}
}

// TestAPartlyEmptiedDiscardPromptTrimsAndClampsItsBounds is the other
// half of the prune, and CR 701.8a's "as many as you can": a discard of
// two out of a hand that has lost two of its three cards is a discard
// of one, and the clamped answer is one the resolver takes.
func TestAPartlyEmptiedDiscardPromptTrimsAndClampsItsBounds(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, g.Seats[1].ID, "Mind Rot")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	for _, name := range []string{"One", "Two", "Three"} {
		handCardForDiscard(t, g, me, name)
	}

	run := discardRunOverOneSeat(t, g, DiscardPrompt{Player: me.ID, Source: source, N: 2})
	c := discardPromptFor(g, me.ID)
	if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("setup: expected a [2,2] prompt, got %v", c)
	}

	// Two of the three candidates are pitched by something else.
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(me.ID, 2); err != nil {
			t.Fatalf("DiscardRandomForEffect: %v", err)
		}
	})

	c = discardPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("a candidate is still in the hand; the prompt must not be withdrawn")
	}
	if len(c.ChooseCards) != 1 {
		t.Errorf("%d candidates after the prune, want 1", len(c.ChooseCards))
	}
	if c.ChooseMin != 1 || c.ChooseMax != 1 {
		t.Errorf("bounds [%d,%d] after the prune, want [1,1] — a floor no remaining set can "+
			"reach is the same wedge one card later", c.ChooseMin, c.ChooseMax)
	}
	if c.Count != 1 {
		t.Errorf("Count = %d after the prune, want 1", c.Count)
	}
	left := c.ChooseCards[0]
	if !me.Hand.Contains(left) {
		t.Fatalf("the surviving candidate is not in the hand")
	}

	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{left}); err != nil {
		t.Fatalf("the clamped maximal answer was refused: %v", err)
	}
	if run.ran != 1 {
		t.Fatalf("the continuation ran %d times, want 1", run.ran)
	}
	if got := run.got.By(me.ID); len(got) != 1 || got[0] != left {
		t.Errorf("the run reports %v, want the one card the prompt could still take", got)
	}
}

// --- the same prune at the other zones --------------------------------

// queueGraveyardPick puts a "choose from your graveyard" prompt up over
// `cards`, the shape Skullwinder and the batch-38 helpers queue.
func queueGraveyardPick(t *testing.T, g *Game, chooser uuid.UUID, source uuid.UUID, cards []uuid.UUID) (uuid.UUID, *int) {
	t.Helper()
	ran := 0
	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  chooser,
			Source:   source,
			Question: "Choose a card in your graveyard",
			Cards:    cards,
			Min:      1,
			Max:      1,
			Zone:     ZoneGraveyard,
			Then:     func(*Game, []uuid.UUID) error { ran++; return nil },
		})
	})
	if id == uuid.Nil {
		t.Fatal("setup: the graveyard pick was not queued")
	}
	return id, &ran
}

// TestAGraveyardPickIsWithdrawnWhenItsCandidatesAreExiled — the same
// prune, one zone over. A pick over a graveyard that a Rest in Peace or
// a Bojuka Bog empties mid-prompt had exactly the discard's hole.
func TestAGraveyardPickIsWithdrawnWhenItsCandidatesAreExiled(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Regrowth")
	var binned []uuid.UUID
	g.WithWriteLock(func() {
		for _, name := range []string{"Buried One", "Buried Two"} {
			c := NewCard(name, me.ID)
			c.TypeLine = "Sorcery"
			me.Graveyard.PushTop(c)
			binned = append(binned, c.InstanceID)
		}
	})

	id, ran := queueGraveyardPick(t, g, me.ID, source, binned)

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(binned[0]); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})
	c := findChoice(g, id)
	if c == nil {
		t.Fatal("one candidate is still in the graveyard; the prompt must not be withdrawn")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != binned[1] {
		t.Errorf("candidates after the prune = %v, want only the card still buried", c.ChooseCards)
	}

	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(binned[1]); err != nil {
			t.Fatalf("ExileCardForEffect: %v", err)
		}
	})
	if findChoice(g, id) != nil {
		t.Error("the prompt survives a graveyard with none of its candidates in it")
	}
	// A pick that is no run's leg has no continuation to settle — the
	// departure table's existing answer for the kind, unchanged here.
	if *ran != 0 {
		t.Errorf("the withdrawn pick ran its continuation %d times", *ran)
	}
	assertTableIsFree(t, g)
}

// TestALibraryPickIsPrunedWhenACandidateLeavesTheLibrary — the third
// zone (Genesis Wave's "put any number of them onto the battlefield",
// the reveal-and-take helpers).
//
// A mid-prompt SHUFFLE is NOT what strands one of these, which is worth
// recording because it is the case the issue asked about: a shuffle
// reorders a library and takes nothing out of it, so every frozen
// candidate is still in the zone the answer is re-checked against and
// the pick stays legal (the ORDER it was chosen in is a different
// question, and not one this prompt asks). What strands it is a
// candidate LEAVING — an exile, a mill, an opponent's tuck.
func TestALibraryPickIsPrunedWhenACandidateLeavesTheLibrary(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Genesis Wave")
	if me.Library.Size() < 2 {
		t.Fatalf("setup: library holds %d cards", me.Library.Size())
	}
	top := []uuid.UUID{
		me.Library.Cards[me.Library.Size()-1].InstanceID,
		me.Library.Cards[me.Library.Size()-2].InstanceID,
	}

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  me.ID,
			Source:   source,
			Question: "Choose a card to put onto the battlefield",
			Cards:    top,
			Min:      1,
			Max:      1,
			Zone:     ZoneLibrary,
			Then:     func(*Game, []uuid.UUID) error { return nil },
		})
	})
	if id == uuid.Nil {
		t.Fatal("setup: the library pick was not queued")
	}

	// A shuffle moves the cards and takes none of them out: the pick is
	// untouched.
	if err := g.ShuffleLibrary(me.ID); err != nil {
		t.Fatalf("ShuffleLibrary: %v", err)
	}
	if c := findChoice(g, id); c == nil || len(c.ChooseCards) != 2 {
		t.Fatalf("a shuffle removed nothing from the library; the pick must be untouched (%v)", c)
	}

	g.WithWriteLock(func() {
		for _, cardID := range top {
			if err := g.ExileCardForEffect(cardID); err != nil {
				t.Fatalf("ExileCardForEffect: %v", err)
			}
		}
	})
	if findChoice(g, id) != nil {
		t.Error("the prompt survives a library with none of its candidates in it")
	}
	assertTableIsFree(t, g)
}

// TestAPickWithNoZoneOnItsFrameIsLeftAlone — the prune's boundary. A
// choose-cards prompt with no Zone re-checks nothing on submit (the
// continuation owns what a pick means), so there is no live list for
// the engine to be right about and the candidates stay exactly as the
// effect froze them. Fact or Fiction's pile split is the shape.
func TestAPickWithNoZoneOnItsFrameIsLeftAlone(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	source := departureTestSource(g, me.ID, "Fact or Fiction")
	var cards []uuid.UUID
	g.WithWriteLock(func() {
		for _, name := range []string{"Pile One", "Pile Two"} {
			c := NewCard(name, me.ID)
			c.TypeLine = "Sorcery"
			me.Graveyard.PushTop(c)
			cards = append(cards, c.InstanceID)
		}
	})

	var id uuid.UUID
	g.WithWriteLock(func() {
		id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
			Chooser:  me.ID,
			Source:   source,
			Question: "Separate those cards into two piles",
			Cards:    cards,
			Then:     func(*Game, []uuid.UUID) error { return nil },
		})
	})

	g.WithWriteLock(func() {
		for _, cardID := range cards {
			if err := g.ExileCardForEffect(cardID); err != nil {
				t.Fatalf("ExileCardForEffect: %v", err)
			}
		}
	})

	c := findChoice(g, id)
	if c == nil {
		t.Fatal("a zoneless pick was withdrawn; nothing about its answer went stale")
	}
	if len(c.ChooseCards) != len(cards) {
		t.Errorf("%d candidates, want the %d the effect froze", len(c.ChooseCards), len(cards))
	}
}
