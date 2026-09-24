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

// cardSetPickCall records what a card-set pick's continuation was run
// with: how many times, and the last set of picks it was handed.
type cardSetPickCall struct {
	ran int
	got []uuid.UUID
}

// queueGraveyardPick puts a "choose from your graveyard" prompt up over
// `cards`, the shape Skullwinder and the batch-38 helpers queue.
func queueGraveyardPick(t *testing.T, g *Game, chooser uuid.UUID, source uuid.UUID, cards []uuid.UUID) (uuid.UUID, *cardSetPickCall) {
	t.Helper()
	call := &cardSetPickCall{}
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
			Then: func(_ *Game, picked []uuid.UUID) error {
				call.ran++
				call.got = picked
				return nil
			},
		})
	})
	if id == uuid.Nil {
		t.Fatal("setup: the graveyard pick was not queued")
	}
	return id, call
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

	id, call := queueGraveyardPick(t, g, me.ID, source, binned)

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
	// #1225: a pick that is no run's leg has no run to settle, but it
	// still holds the REST OF THE CARD in its frame, and the drop runs
	// it with nothing picked — the same closure the empty-candidate
	// path at queue time would have run with the same argument.
	if call.ran != 1 {
		t.Errorf("the withdrawn pick ran its continuation %d times, want 1 — "+
			"the rest of the card went with the question", call.ran)
	}
	if len(call.got) != 0 {
		t.Errorf("the continuation was handed %v; a drop chooses nothing on the chooser's behalf", call.got)
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

// TestASeatLeavingEmptiesEveryPickAtItsHandInOnePass covers the third
// call site — the departure sweep — and the reason the sweep examines
// the whole queue before it drops anything.
//
// removeObjectsOwnedByLocked takes the departed player's cards out of
// every zone DIRECTLY (leaving the game is not a zone change), so the
// exit prune never sees them, and it empties two survivors' prompts in
// the same breath. Dropping the first runs its drop action, which may
// rewrite the queue the loop is standing in; the second must still go.
func TestASeatLeavingEmptiesEveryPickAtItsHandInOnePass(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	leaver := g.Seats[1]
	theirHand := handCardOf(t, leaver, 2)

	ids := make([]uuid.UUID, 0, 2)
	for _, chooser := range []*Player{g.Seats[0], g.Seats[2]} {
		source := departureTestSource(g, chooser.ID, "Thoughtseize")
		var id uuid.UUID
		g.WithWriteLock(func() {
			id = g.QueueChooseCardsForEffect(ChooseCardsPrompt{
				Chooser:    chooser.ID,
				FromPlayer: leaver.ID,
				Source:     source,
				Question:   "Choose a card from that player's hand",
				Cards:      theirHand,
				Min:        1,
				Max:        1,
				Zone:       ZoneHand,
				Then:       func(*Game, []uuid.UUID) error { return nil },
			})
		})
		if id == uuid.Nil {
			t.Fatal("setup: the pick was not queued")
		}
		ids = append(ids, id)
	}

	if err := g.Concede(leaver.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	for i, id := range ids {
		if findChoice(g, id) != nil {
			t.Errorf("pick %d survives a hand that left the game with its owner", i)
		}
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

// --- #1263: the two kinds this prune must never withdraw --------------
//
// untap_choice and entry_reveal_from_hand both carry the choose-cards
// payload (isCardSetPickKind) and both set a real Zone on their frame
// (ZoneBattlefield, ZoneHand), so — before this fix — the loop above
// examined them exactly like a discard prompt and could WITHDRAW one
// whose candidates all left. Neither kind's departure row can survive
// that: untap_choice's is dropDiscard, which strands the untap step
// forever, and entry_reveal_from_hand's paused replacementResume is
// only settled by dropChoicesForPlayerLocked (a player's departure),
// never by the plain withdrawal dropChoiceLocked performs here. The
// doc comment on pruneCardSetChoicesLocked has always said both kinds
// are "left alone" — these two tests are what makes that true rather
// than aspirational.

// TestAnUntapChoiceIsNotWithdrawnWhenItsCandidateLeavesTheBattlefield
// is the untap_choice shape. The floor is 0 (a pure opt-out, no cap),
// so once the fix holds the prompt survives AND the step can still
// complete with nothing left to determine — proof this is not merely
// "not silently dropped" but genuinely still answerable.
func TestAnUntapChoiceIsNotWithdrawnWhenItsCandidateLeavesTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	const tick = "rust-tick"
	seat := g.Seats[0].ID
	id := pushTappedPermanent(g, seat, "Tick", tick, "Artifact Creature", true)
	withCatalogUntapOptOuts(t, optOutSelf(tick))

	g.WithWriteLock(func() { g.performUntapStepLocked(0) })
	prompt := openUntapChoice(t, g)
	if prompt.ChooseMin != 0 {
		t.Fatalf("setup: floor = %d, want 0 (a pure opt-out is never mandatory)", prompt.ChooseMin)
	}

	// The only candidate leaves the battlefield while the
	// determination is still open — an admin move or a concession, per
	// finishUntapStepLocked's own doc comment on why that can happen
	// mid-step.
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
		g.pruneCardSetChoicesLocked()
	})

	if findChoice(g, prompt.ID) == nil {
		t.Fatal("the untap_choice prompt was withdrawn when its only candidate left the battlefield — " +
			"its drop action is dropDiscard, which runs nothing, so the untap step could never exit")
	}

	// Still answerable: the floor is 0, so "nothing left to determine"
	// is a legal answer and the step completes normally.
	if err := g.ResolveUntapChoice(prompt.ID, seat, nil); err != nil {
		t.Fatalf("ResolveUntapChoice with nothing left to choose: %v", err)
	}
	if g.Turn.Step != StepUpkeep {
		t.Errorf("step after the answer = %q, want upkeep — the untap step must still exit", g.Turn.Step)
	}
}

// stubEntryHandRevealLand registers a synthetic land carrying an
// EntryHandReveal replacement — reveal_lands.go's real shape
// (EntersTappedUnlessYouRevealFromHand), rebuilt directly here because
// internal/game cannot import cards/effects and this test needs to
// reach pruneCardSetChoicesLocked, which is unexported.
func stubEntryHandRevealLand(t *testing.T, oracleID string) {
	t.Helper()
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		oracleID: {{
			Watches:         []EventKind{EventZoneMove},
			SelfReplacement: true,
			Label:           "Test Reveal Land",
			PromptQuestion:  "Test Reveal Land — reveal a card so it enters untapped?",
			// Matches excludes the entering land itself — it is still
			// physically in the hand's slice at this pre-push point,
			// and Nil would admit it right alongside the real
			// candidate.
			EntryHandReveal: &EntryHandReveal{Min: 0, Max: 1, Matches: func(c Card) bool { return c.Name == "Island" }},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
				return ev.Kind == RepEventMove && ev.NewZone == ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
			Controller: func(_ *ReplacementEvent, _ *Game, src *Card) uuid.UUID { return src.Controller },
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.EntersTapped = true
				return nil
			},
		}},
	})
}

// TestAnEntryRevealFromHandIsNotWithdrawnWhenItsCandidateLeavesTheHand
// is the entry_reveal_from_hand shape: its floor is always 0 (#1198 —
// "you may reveal", never "you must"), so once the fix holds, the
// prompt survives AND the paused CR 614 entry can still be resumed by
// declining — proof this is not merely "not silently dropped" but
// genuinely still answerable.
func TestAnEntryRevealFromHandIsNotWithdrawnWhenItsCandidateLeavesTheHand(t *testing.T) {
	g := newActiveGame(t)
	const oracleID = "test-reveal-land"
	stubEntryHandRevealLand(t, oracleID)

	active := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { active.Hand.Cards = nil })
	landID := uuid.New()
	active.Hand.PushTop(Card{
		InstanceID: landID, Name: "Test Land", TypeLine: "Land",
		OracleID: oracleID, Owner: active.ID, Controller: active.ID,
	})
	islandID := uuid.New()
	active.Hand.PushTop(Card{
		InstanceID: islandID, Name: "Island", TypeLine: "Basic Land — Island",
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != StepPrecombatMain && g.Turn.Step != StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, landID, CastSpellParams{}); err != nil {
		t.Fatalf("play the land: %v", err)
	}

	var prompt *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceEntryRevealFromHand && c.Chooser == active.ID {
			prompt = c
		}
	}
	if prompt == nil {
		t.Fatal("setup: playing the land queued no entry_reveal_from_hand prompt")
	}
	if len(prompt.ChooseCards) != 1 || prompt.ChooseCards[0] != islandID {
		t.Fatalf("setup: candidates = %v, want just the Island", prompt.ChooseCards)
	}
	if _, ok := battlefieldCardByID(g, landID); ok {
		t.Fatal("setup: the land reached the battlefield before the choice was made")
	}

	// The only candidate leaves the hand while the entry is still
	// paused — some other effect discards or exiles it. The entering
	// land itself is still physically sitting in this same hand
	// (pre-push — nothing has moved yet, see
	// TestRevealLandPromptPausesTheEntry's twin assertion in
	// cards/effects), so only the Island comes out.
	g.WithWriteLock(func() {
		if _, err := active.Hand.Remove(islandID); err != nil {
			t.Fatalf("Hand.Remove(island): %v", err)
		}
		g.pruneCardSetChoicesLocked()
	})

	found := findChoice(g, prompt.ID)
	if found == nil {
		t.Fatal("the entry_reveal_from_hand prompt was withdrawn when its only candidate left the hand — " +
			"the paused CR 614 entry is now stranded forever, since dropChoiceLocked does not settle a " +
			"replacementResume the way a player's departure does")
	}

	// Still answerable: the floor is 0, so declining ("reveal
	// nothing", the only real answer left) resumes the entry.
	if err := g.ResolveEntryRevealFromHand(found.ID, active.ID, nil); err != nil {
		t.Fatalf("ResolveEntryRevealFromHand with nothing left to reveal: %v", err)
	}
	bf, ok := battlefieldCardByID(g, landID)
	if !ok {
		t.Fatal("declining the reveal did not resume the entry — the land never reached the battlefield")
	}
	if !bf.Tapped {
		t.Error("the land should have entered tapped — the reveal was declined")
	}
}
