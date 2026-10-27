package game

import (
	"testing"

	"github.com/google/uuid"
)

// discard_entry_test.go — ADR 0153 PR 1 (#1861): a discard an opponent
// causes, and a discard that a replacement settles on the battlefield.
//
// Every case drives a synthetic card whose own replacement is the
// Obstinate Baloth shape ("if a spell or ability an opponent controls
// causes you to discard this card, put it onto the battlefield instead
// of putting it into your graveyard"), so the engine is pinned without
// the catalog. The real cards land in ADR 0153 PR 2.

const (
	discardEntryOracle  = "test-discard-onto-the-battlefield"
	discardEntryScales  = "test-discard-entry-hardened-scales"
	discardEntryLabel   = "test: onto the battlefield instead"
	discardEntryTappedA = "test: enters tapped (a)"
	discardEntryTappedB = "test: enters tapped (b)"
)

// balothReplacement is the family's replacement, with optional counters
// for Dodecapod.
func balothReplacement(counters map[string]int) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventDiscardCard},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, src *Card) bool {
			return ev.Kind == RepEventDiscard && src != nil && ev.CardID == src.InstanceID &&
				ev.NewZone == ZoneGraveyard && DiscardCausedByOpponent(ev)
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.NewZone = ZoneBattlefield
			ev.NewZoneOwner = uuid.Nil
			for name, n := range counters {
				ev.AddCounterAtETB(name, n)
			}
			return nil
		},
		Controller: func(ev *ReplacementEvent, _ *Game, _ *Card) uuid.UUID { return ev.DiscardPlayer },
		Label:      discardEntryLabel,
	}
}

// boundTo turns a card's own replacement into a test-registered one
// that reads `card` as its source.
func boundTo(card uuid.UUID, r ReplacementEffect) ReplacementEffect {
	src := func() *Card { return &Card{InstanceID: card} }
	applies, replace, controller := r.AppliesTo, r.Replace, r.Controller
	r.AppliesTo = func(ev *ReplacementEvent, g *Game, _ *Card) bool { return applies(ev, g, src()) }
	r.Replace = func(ev *ReplacementEvent, g *Game, _ *Card) error { return replace(ev, g, src()) }
	r.Controller = func(ev *ReplacementEvent, g *Game, _ *Card) uuid.UUID { return controller(ev, g, src()) }
	return r
}

// stubBaloth wires the family's replacement onto discardEntryOracle,
// alongside any other catalog replacements the test needs.
func stubBaloth(t *testing.T, counters map[string]int, extra map[string][]ReplacementEffect) {
	t.Helper()
	byKey := map[string][]ReplacementEffect{discardEntryOracle: {balothReplacement(counters)}}
	for k, v := range extra {
		byKey[k] = append(byKey[k], v...)
	}
	stubCatalogReplacements(t, byKey)
}

// seedBaloth puts a 4/4 creature card carrying the replacement into
// p's hand, with an empty hand otherwise.
func seedBaloth(t *testing.T, g *Game, p *Player) uuid.UUID {
	t.Helper()
	c := NewCard("Test Baloth", p.ID)
	c.OracleID = discardEntryOracle
	c.TypeLine = "Creature — Beast"
	c.ManaCost = "{2}{G}{G}"
	c.Power, c.Toughness = 4, 4
	g.WithWriteLock(func() {
		p.Hand.Cards = nil
		p.Hand.PushTop(c)
	})
	return c.InstanceID
}

// discardUnder discards `cards` from p's hand with `cause`, while a
// stack item controlled by `itemController` (uuid.Nil for none) is
// resolving, and records what the discard's "then" was told.
func discardUnder(t *testing.T, g *Game, p *Player, itemController uuid.UUID, cause DiscardCause, cards ...uuid.UUID) *discardThen {
	t.Helper()
	got := &discardThen{}
	g.WithWriteLock(func() {
		if itemController != uuid.Nil {
			resolvingAbilityOf(g, itemController)
		}
		err := g.discardCardsLocked(p.ID, cards, discardOptions{
			cause: cause,
			keyed: hookTail(func(_ *Game, o TailOutcome) error {
				got.runs++
				got.landed = append([]uuid.UUID(nil), o.IDs...)
				return nil
			}),
		})
		g.resolving = nil // play moves on before any prompt is answered
		if err != nil {
			t.Fatalf("discardCardsLocked: %v", err)
		}
	})
	return got
}

// assertDiscardOntoBattlefield is assertOneDiscardEvent for a card a
// replacement put onto the battlefield. One EventDiscardCard, hand →
// battlefield. The entry's own hand → battlefield EventZoneMove is
// expected: the card enters exactly as any card put from a hand does
// (ADR 0153 §2), and no graveyard watcher (Syr Konrad) can read it.
func assertDiscardOntoBattlefield(t *testing.T, w *discardWatcher, actor, card uuid.UUID) {
	t.Helper()
	if len(w.discards) != 1 {
		t.Fatalf("EventDiscardCard x %d, want 1", len(w.discards))
	}
	ev := w.discards[0]
	if ev.Actor != actor || ev.CardID != card || ev.OldZone != ZoneHand || ev.NewZone != ZoneBattlefield {
		t.Errorf("EventDiscardCard = %s %s → %s by %s, want the card hand → battlefield by %s",
			ev.CardID, ev.OldZone, ev.NewZone, ev.Actor, actor)
	}
	for _, zm := range w.zoneMoves {
		if zm.CardID == card && zm.NewZone == ZoneGraveyard {
			t.Error("the card was also moved to a graveyard")
		}
	}
}

type discardThen struct {
	runs   int
	landed []uuid.UUID
}

// eventIndex is the index in g.Events of the first event of `kind`
// about `card` at or after `from`, or -1.
func deEventIndex(g *Game, from int, kind EventKind, card uuid.UUID) int {
	for i := from; i < len(g.Events); i++ {
		if g.Events[i].Kind == kind && g.Events[i].CardID == card {
			return i
		}
	}
	return -1
}

// --- §1: the replacement event carries the cause ----------------------

// TestADiscardReplacementEventCarriesItsSourceAndCauseController: the
// event a discard replacement sees names the card that asked (Source,
// which its comment always claimed) and the resolving item's
// controller, and only for an effect.
func TestADiscardReplacementEventCarriesItsSourceAndCauseController(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cause DiscardCause
		item  bool
		want  bool // DiscardCauseController is the opponent
	}{
		{"an opponent's effect", DiscardCauseEffect, true, true},
		{"a cost paid while the opponent's item resolves", DiscardCauseCost, true, false},
		{"the cleanup step", DiscardCauseCleanup, false, false},
		{"an effect with nothing resolving", DiscardCauseEffect, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p, opp := g.Seats[1], g.Seats[0]
			g.WithWriteLock(func() { p.Hand.Cards = nil })
			card := handCardForDiscard(t, g, p, "Pitch")
			source := uuid.New()

			var seen *ReplacementEvent
			g.WithWriteLock(func() {
				g.RegisterReplacementForTest(ReplacementEffect{
					Watches: []EventKind{EventDiscardCard},
					AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
						cp := *ev
						seen = &cp
						return false
					},
					Label: "observer",
				})
				if tc.item {
					resolvingAbilityOf(g, opp.ID)
				}
				if err := g.discardCardsLocked(p.ID, []uuid.UUID{card}, discardOptions{cause: tc.cause, source: source}); err != nil {
					t.Fatalf("discardCardsLocked: %v", err)
				}
				g.resolving = nil
			})
			if seen == nil {
				t.Fatal("no discard window opened")
			}
			if seen.Source != source {
				t.Errorf("ReplacementEvent.Source = %s, want the discard's source %s", seen.Source, source)
			}
			if got := seen.DiscardCauseController == opp.ID; got != tc.want {
				t.Errorf("DiscardCauseController = %s, want opponent: %v", seen.DiscardCauseController, tc.want)
			}
			if !tc.want && seen.DiscardCauseController != uuid.Nil {
				t.Errorf("DiscardCauseController = %s, want uuid.Nil for a %s discard", seen.DiscardCauseController, tc.cause)
			}
			if DiscardCausedByOpponent(seen) != tc.want {
				t.Errorf("DiscardCausedByOpponent = %v, want %v", !tc.want, tc.want)
			}
		})
	}
}

// --- §2: the battlefield handoff -----------------------------------------

// TestAnOpponentsDiscardPutsTheCardOntoTheBattlefield is the family's
// whole point: it enters under its OWNER's control, summoning sick,
// EventETB and then EventDiscardCard (with the true zone and the cause)
// are emitted, and the discard's "then" counts it as discarded.
func TestAnOpponentsDiscardPutsTheCardOntoTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	p, opp := g.Seats[1], g.Seats[0]
	stubBaloth(t, nil, nil)
	card := seedBaloth(t, g, p)
	before := len(g.Events)

	then := discardUnder(t, g, p, opp.ID, DiscardCauseEffect, card)

	assertOnlyIn(t, card, g.Battlefield, p.Hand, p.Graveyard)
	c := battlefieldCardForTest(t, g, card)
	if c.Controller != p.ID {
		t.Errorf("controller = %s, want the owner %s (CR 110.2)", c.Controller, p.ID)
	}
	if !c.SummonedThisTurn {
		t.Error("the creature entered without summoning sickness (CR 302.6)")
	}
	etb := deEventIndex(g, before, EventETB, card)
	dis := deEventIndex(g, before, EventDiscardCard, card)
	if etb < 0 || dis < 0 {
		t.Fatalf("EventETB at %d, EventDiscardCard at %d; want both", etb, dis)
	}
	if etb > dis {
		t.Error("EventDiscardCard came before EventETB (owner decision 3: ETB first)")
	}
	ev := g.Events[dis]
	if ev.OldZone != ZoneHand || ev.NewZone != ZoneBattlefield || ev.Actor != p.ID {
		t.Errorf("EventDiscardCard %s → %s by %s, want hand → battlefield by the owner", ev.OldZone, ev.NewZone, ev.Actor)
	}
	if ev.Cause != MoveCauseEffect || ev.CauseController != opp.ID || ev.DiscardCause != DiscardCauseEffect {
		t.Errorf("EventDiscardCard cause = %q/%s/%q, want effect/opponent/effect", ev.Cause, ev.CauseController, ev.DiscardCause)
	}
	if !DiscardCausedByOpponentOf(ev, p.ID) {
		t.Error("DiscardCausedByOpponentOf is false for an opponent's discard")
	}
	if n := deCountEvents(g, before, EventDiscardCard, card); n != 1 {
		t.Errorf("EventDiscardCard x %d, want 1", n)
	}
	if then.runs != 1 || len(then.landed) != 1 || then.landed[0] != card {
		t.Errorf(`"then" ran %d time(s) with %v; want once, counting the card as discarded`, then.runs, then.landed)
	}
}

func deCountEvents(g *Game, from int, kind EventKind, card uuid.UUID) int {
	n := 0
	for i := from; i < len(g.Events); i++ {
		if g.Events[i].Kind == kind && g.Events[i].CardID == card {
			n++
		}
	}
	return n
}

// TestOnlyAnOpponentsEffectDiscardIsReplaced: the owner's own spell, a
// cost (the owner's echo or cumulative upkeep discard included) and the
// cleanup discard leave the card in the graveyard, and it was still
// discarded. Painful Quandary is an opponent's EFFECT discard (owner
// decision 4), the first row.
func TestOnlyAnOpponentsEffectDiscardIsReplaced(t *testing.T) {
	for _, tc := range []struct {
		name    string
		item    func(g *Game) uuid.UUID
		cause   DiscardCause
		wantBF  bool
		wantOpp bool
	}{
		{"an opponent's spell", func(g *Game) uuid.UUID { return g.Seats[0].ID }, DiscardCauseEffect, true, true},
		{"the owner's own spell", func(g *Game) uuid.UUID { return g.Seats[1].ID }, DiscardCauseEffect, false, false},
		{"a cost while an opponent's item resolves", func(g *Game) uuid.UUID { return g.Seats[0].ID }, DiscardCauseCost, false, false},
		{"the owner's own cost (echo)", func(g *Game) uuid.UUID { return g.Seats[1].ID }, DiscardCauseCost, false, false},
		{"the cleanup step", func(*Game) uuid.UUID { return uuid.Nil }, DiscardCauseCleanup, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p := g.Seats[1]
			stubBaloth(t, nil, nil)
			card := seedBaloth(t, g, p)
			w := watchDiscards(g)

			then := discardUnder(t, g, p, tc.item(g), tc.cause, card)

			if tc.wantBF {
				assertOnlyIn(t, card, g.Battlefield, p.Hand, p.Graveyard)
				assertDiscardOntoBattlefield(t, w, p.ID, card)
			} else {
				assertOnlyIn(t, card, p.Graveyard, p.Hand, g.Battlefield)
				assertOneDiscardEvent(t, w, p.ID, card, ZoneGraveyard)
			}
			if got := DiscardCausedByOpponentOf(w.discards[0], p.ID); got != tc.wantOpp {
				t.Errorf("DiscardCausedByOpponentOf = %v, want %v", got, tc.wantOpp)
			}
			if then.runs != 1 || len(then.landed) != 1 {
				t.Errorf(`"then" ran %d time(s) with %v; want once with the card`, then.runs, then.landed)
			}
		})
	}
}

// TestTheStackItemsControllerDecidesNotTheSourceCards is CR 113.8: an
// ability whose source card the OWNER controls but whose stack item an
// opponent controls is the opponent's, and the reverse is not.
func TestTheStackItemsControllerDecidesNotTheSourceCards(t *testing.T) {
	for _, tc := range []struct {
		name      string
		itemSeat  int
		cardSeat  int
		wantEnter bool
	}{
		{"opponent's item, owner's card", 0, 1, true},
		{"owner's item, opponent's card", 1, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p := g.Seats[1]
			stubBaloth(t, nil, nil)
			card := seedBaloth(t, g, p)
			srcCard := pushBear(g, g.Seats[tc.cardSeat].ID)

			g.WithWriteLock(func() {
				item := resolvingAbilityOf(g, g.Seats[tc.itemSeat].ID)
				item.SourceCardID = srcCard
				if err := g.discardCardsLocked(p.ID, []uuid.UUID{card}, discardOptions{cause: DiscardCauseEffect, source: srcCard}); err != nil {
					t.Fatalf("discardCardsLocked: %v", err)
				}
				g.resolving = nil
			})

			if tc.wantEnter {
				assertOnlyIn(t, card, g.Battlefield, p.Hand, p.Graveyard)
			} else {
				assertOnlyIn(t, card, p.Graveyard, p.Hand, g.Battlefield)
			}
		})
	}
}

// --- madness and CR 616 ------------------------------------------------

// TestMadnessAndTheBattlefieldReplacementAreOrderedByTheDiscarder: both
// apply to the same discard, the discarding player orders them, and the
// one applied second no longer applies once the first has moved the
// destination off the graveyard (CR 614.5). The battlefield-first order
// also proves the RESUMED path reaches the handoff.
func TestMadnessAndTheBattlefieldReplacementAreOrderedByTheDiscarder(t *testing.T) {
	for _, tc := range []struct {
		name  string
		first string
		want  func(g *Game, p *Player) *Zone
	}{
		{"battlefield first", discardEntryLabel, func(g *Game, _ *Player) *Zone { return g.Battlefield }},
		{"madness first", MadnessReplacementLabel, func(g *Game, _ *Player) *Zone { return g.Exile }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p, opp := g.Seats[1], g.Seats[0]
			// Both registered as labelled test effects bound to the
			// card, so the ordering prompt can be answered by label (a
			// hand card's catalog replacement carries no label lookup).
			card := seedBaloth(t, g, p)
			g.WithWriteLock(func() {
				g.RegisterReplacementForTest(boundTo(card, balothReplacement(nil)))
				g.RegisterReplacementForTest(boundTo(card, MadnessReplacement()))
			})
			w := watchDiscards(g)

			then := discardUnder(t, g, p, opp.ID, DiscardCauseEffect, card)

			prompt := replacementOrderPrompt(t, g)
			if prompt.Chooser != p.ID {
				t.Fatalf("chooser = %s, want the discarding player (CR 616.1)", prompt.Chooser)
			}
			if then.runs != 0 || !p.Hand.Contains(card) {
				t.Fatal("the discard moved or finished before the order was chosen")
			}
			second := MadnessReplacementLabel
			if tc.first == MadnessReplacementLabel {
				second = discardEntryLabel
			}
			if err := g.ResolveReplacementOrder(prompt.ID, p.ID, orderByLabel(t, g, prompt, tc.first, second)); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}

			want := tc.want(g, p)
			assertOnlyIn(t, card, want, p.Hand, p.Graveyard)
			if want == g.Exile && g.Battlefield.Contains(card) {
				t.Error("the card is on the battlefield as well as in exile")
			}
			if want.Kind == ZoneBattlefield {
				assertDiscardOntoBattlefield(t, w, p.ID, card)
			} else {
				assertOneDiscardEvent(t, w, p.ID, card, want.Kind)
			}
			if then.runs != 1 || len(then.landed) != 1 {
				t.Errorf(`"then" ran %d time(s) with %v; want once with the card`, then.runs, then.landed)
			}
		})
	}
}

// --- §3: the entry's own window --------------------------------------

// enterTapped is an entry replacement on the card: "it enters tapped".
func enterTapped(card uuid.UUID, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventZoneMove},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventMove && ev.CardID == card && ev.NewZone == ZoneBattlefield && !ev.EntersTapped
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.EntersTapped = true
			return nil
		},
		Controller: func(ev *ReplacementEvent, _ *Game, _ *Card) uuid.UUID { return ev.Actor },
		Label:      label,
	}
}

// TestAPausedEntryHoldsTheRestOfTheBatchAndUndoes: the entry window
// asks a question (two enters-tapped effects to order). The card stays
// in hand, the second card of the batch and the discard's "then" wait,
// and both finish after the answer. An undo taken with the question
// open puts both cards back and replays the same way.
func TestAPausedEntryHoldsTheRestOfTheBatchAndUndoes(t *testing.T) {
	g := newActiveGame(t)
	p, opp := g.Seats[1], g.Seats[0]
	stubBaloth(t, nil, nil)
	card := seedBaloth(t, g, p)
	spare := handCardForDiscard(t, g, p, "Spare")
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(enterTapped(card, discardEntryTappedA))
		g.RegisterReplacementForTest(enterTapped(card, discardEntryTappedB))
	})

	then := discardUnder(t, g, p, opp.ID, DiscardCauseEffect, card, spare)

	replacementOrderPrompt(t, g)
	if !p.Hand.Contains(card) || !p.Hand.Contains(spare) {
		t.Fatal("a card left the hand while the entry's question was open")
	}
	if then.runs != 0 {
		t.Fatal(`the discard's "then" ran while the entry was paused`)
	}
	snap := g.Clone()

	answer := func(g *Game) {
		t.Helper()
		pc := replacementOrderPrompt(t, g)
		if err := g.ResolveReplacementOrder(pc.ID, pc.Chooser, orderByLabel(t, g, pc, discardEntryTappedA, discardEntryTappedB)); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
	}
	check := func(label string) {
		t.Helper()
		p := g.Seats[1]
		assertOnlyIn(t, card, g.Battlefield, p.Hand, p.Graveyard)
		assertOnlyIn(t, spare, p.Graveyard, p.Hand, g.Battlefield)
		if !ebTapped(g, card) {
			t.Errorf("%s: the entry's enters-tapped did not apply", label)
		}
		if then.runs != 1 || len(then.landed) != 2 {
			t.Errorf(`%s: "then" ran %d time(s) with %v; want once with both cards`, label, then.runs, then.landed)
		}
		if len(g.PendingChoices) != 0 {
			t.Errorf("%s: %d prompt(s) left open", label, len(g.PendingChoices))
		}
	}
	answer(g)
	check("live")

	then.runs, then.landed = 0, nil
	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if !g.Seats[1].Hand.Contains(card) || !g.Seats[1].Hand.Contains(spare) {
		t.Fatal("the undo did not put both cards back in the hand")
	}
	answer(g)
	check("replay")
}

// TestAnEntryRedirectedToExileIsStillADiscard is ADR 0098 Decision 4
// meeting ADR 0153: the entry window sends the card to exile. It ends
// there, EventDiscardCard names exile, and the card counts as
// discarded.
func TestAnEntryRedirectedToExileIsStillADiscard(t *testing.T) {
	g := newActiveGame(t)
	p, opp := g.Seats[1], g.Seats[0]
	stubBaloth(t, nil, nil)
	card := seedBaloth(t, g, p)
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventZoneMove},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventMove && ev.CardID == card && ev.NewZone == ZoneBattlefield
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.NewZone = ZoneExile
				ev.NewZoneOwner = uuid.Nil
				return nil
			},
			Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return opp.ID },
			Label:      "test: exile it instead",
		})
	})
	w := watchDiscards(g)

	then := discardUnder(t, g, p, opp.ID, DiscardCauseEffect, card)

	assertOnlyIn(t, card, g.Exile, p.Hand, p.Graveyard, g.Battlefield)
	if len(w.discards) != 1 || w.discards[0].NewZone != ZoneExile || w.discards[0].OldZone != ZoneHand {
		t.Fatalf("EventDiscardCard = %+v, want one, hand → exile", w.discards)
	}
	if then.runs != 1 || len(then.landed) != 1 || then.landed[0] != card {
		t.Errorf(`"then" ran %d time(s) with %v; want once, counting the card`, then.runs, then.landed)
	}
	if c, ok := g.LookupCardForEffect(card); ok && c.Controller != p.ID {
		t.Errorf("the exiled card kept the entry's controller stamp: %s", c.Controller)
	}
}

// TestTheDiscardReplacementsCountersRideTheEntry is Dodecapod: two
// +1/+1 counters, placed as the entry's own, so Hardened Scales adds
// its one.
func TestTheDiscardReplacementsCountersRideTheEntry(t *testing.T) {
	for _, tc := range []struct {
		name   string
		scales bool
		want   int
	}{
		{"alone", false, 2},
		{"under Hardened Scales", true, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newActiveGame(t)
			p, opp := g.Seats[1], g.Seats[0]
			stubBaloth(t, map[string]int{CounterPlusOne: 2}, map[string][]ReplacementEffect{
				discardEntryScales: {counterAdder("Test Hardened Scales", 1)},
			})
			if tc.scales {
				g.WithWriteLock(func() { pushReplacementSource(g, discardEntryScales, p.ID) })
			}
			card := seedBaloth(t, g, p)

			discardUnder(t, g, p, opp.ID, DiscardCauseEffect, card)

			assertOnlyIn(t, card, g.Battlefield, p.Hand, p.Graveyard)
			if got := counters(t, g, card, CounterPlusOne); got != tc.want {
				t.Errorf("+1/+1 counters = %d, want %d", got, tc.want)
			}
		})
	}
}

// TestADiscardWhoseCardLeftTheHandMovesNothing: the card is no longer
// in hand by the time the handoff runs (it moved while a prompt about
// the discard was open). Nothing moves, no discard is announced, and
// the rest of the batch still runs.
func TestADiscardWhoseCardLeftTheHandMovesNothing(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1]
	stubBaloth(t, nil, nil)
	card := seedBaloth(t, g, p)
	tailRuns := 0
	before := len(g.Events)
	g.WithWriteLock(func() {
		c, err := p.Hand.Remove(card)
		if err != nil {
			t.Fatalf("remove: %v", err)
		}
		p.Graveyard.PushTop(c)
		err = g.executeZoneRouteLocked(&ReplacementEvent{
			Kind:          RepEventDiscard,
			CardID:        card,
			OldZone:       ZoneHand,
			NewZone:       ZoneBattlefield,
			DiscardPlayer: p.ID,
			DiscardCause:  DiscardCauseEffect,
			zoneRoute: &zoneRoute{
				CardID: card, Dst: ZoneGraveyard, DstOwner: p.ID, Actor: p.ID, Discard: true,
				discardLegs: &discardBatchRun{Player: p.ID, Leg: card, Then: snapshotKeyedTail(hookTail(func(*Game, TailOutcome) error {
					tailRuns++
					return nil
				}))},
			},
		})
		if err != nil {
			t.Fatalf("executeZoneRouteLocked: %v", err)
		}
	})
	assertOnlyIn(t, card, p.Graveyard, p.Hand, g.Battlefield)
	if tailRuns != 1 {
		t.Errorf("the batch's continuation ran %d times, want 1", tailRuns)
	}
	if n := deCountEvents(g, before, EventDiscardCard, card); n != 0 {
		t.Errorf("EventDiscardCard x %d for a card that never left the hand by this discard", n)
	}
}

// TestDiscardCausedByOpponentOfReadsOnlyTheDiscardingPlayer pins the
// triggered half's predicate on hand-built events.
func TestDiscardCausedByOpponentOfReadsOnlyTheDiscardingPlayer(t *testing.T) {
	me, opp := uuid.New(), uuid.New()
	base := Event{Kind: EventDiscardCard, Actor: me, Cause: MoveCauseEffect, CauseController: opp}
	if !DiscardCausedByOpponentOf(base, me) {
		t.Error("an opponent's effect discard of mine: want true")
	}
	if DiscardCausedByOpponentOf(base, opp) {
		t.Error("asked about the opponent, who did not discard: want false")
	}
	own := base
	own.CauseController = me
	if DiscardCausedByOpponentOf(own, me) {
		t.Error("my own effect: want false")
	}
	cost := base
	cost.Cause = MoveCauseCost
	if DiscardCausedByOpponentOf(cost, me) {
		t.Error("a cost: want false")
	}
	move := base
	move.Kind = EventZoneMove
	if DiscardCausedByOpponentOf(move, me) {
		t.Error("not a discard: want false")
	}
}
