package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_narrated_silences_test.go — #1021. #984 wrote the log's 39
// deliberate silences down in one place for the first time, and six of
// them read as gaps rather than decisions: a control change, a special
// action, a cycling, a counter landing, a finished scry or surveil, and
// a Saga chapter or Class level. Each is one arm in projectEvent now,
// and each of these tests is the line it produces.
//
// Three properties every assertion below shares, because they are the
// three the log lives or dies by:
//
//   - the line names PEOPLE and CARDS by name and never a UUID, which
//     assertNoUUID (log_choice_test.go) checks;
//   - a scry or a surveil names no card AT ALL, for anyone;
//   - a value that identifies the card it is about — a counter kind, a
//     printed special-action label — is redacted with the card's name,
//     the way #781 and #984 redact a chosen colour.

// battlefieldCard seeds one permanent both seats know about and
// returns its instance ID. Caller must hold no lock; this takes one.
func battlefieldCard(t *testing.T, g *game.Game, name, typeLine string, controller uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	known := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		known[p.ID] = true
	}
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine,
			Owner: controller, Controller: controller, KnownBy: known,
		})
	})
	return id
}

// --- 1. a control change (CR 613.1b) ---------------------------------

// The strongest of the six: "Ian gained control of Grizzly Bears" is a
// thing a player says out loud, and before this the card simply
// appeared under a different controller on the next frame.
func TestLogNarratesAControlChange(t *testing.T) {
	g := buildActiveGame(t)
	gained, lost := g.Seats[0], g.Seats[1]
	bear := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", lost.ID)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventControlChanged, Actor: gained.ID, Target: lost.ID, CardID: bear,
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogControl)
	if want := "P1 gained control of Grizzly Bears from P2"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Seat != 0 {
		t.Errorf("seat: got %d, want 0 — the seat that GAINED control", entry.Seat)
	}
	if entry.TargetSeat == nil || *entry.TargetSeat != 1 {
		t.Errorf("target_seat: got %v, want seat 1 — the seat that lost it", entry.TargetSeat)
	}
	if entry.CardID != bear.String() {
		t.Errorf("card_id: got %q, want the bear", entry.CardID)
	}
	assertNoUUID(t, entry.Text)
}

// A control change with no losing seat — a token nobody controlled, a
// seat the view no longer carries — still says what happened.
func TestLogNarratesAControlChangeWithNoLosingSeat(t *testing.T) {
	g := buildActiveGame(t)
	gained := g.Seats[0]
	bear := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", gained.ID)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventControlChanged, Actor: gained.ID, CardID: bear})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogControl)
	if want := "P1 gained control of Grizzly Bears"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.TargetSeat != nil {
		t.Errorf("target_seat: got %v, want none", entry.TargetSeat)
	}
	assertNoUUID(t, entry.Text)
}

// --- 2. a special action (CR 116.2) ----------------------------------

// The zone move says a card left a hand for exile; only this says it
// was a foretell, and what it cost. game.EventSpecialAction's doc
// comment claimed exactly this before there was an arm to make it true.
func TestLogNarratesASpecialAction(t *testing.T) {
	g := buildActiveGame(t)
	actor, other := g.Seats[0], g.Seats[1]
	card := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{
			InstanceID: card, Name: "Saw It Coming", TypeLine: "Instant",
			Owner: actor.ID, Controller: actor.ID,
			KnownBy: map[uuid.UUID]bool{actor.ID: true, other.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventSpecialAction, Actor: actor.ID, CardID: card, Label: "Foretell {2}",
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogSpecialAction)
	if want := "P1 used Foretell {2} on Saw It Coming"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Label != "Foretell {2}" {
		t.Errorf("label: got %q, want the printed action", entry.Label)
	}
	if entry.Seat != 0 {
		t.Errorf("seat: got %d, want 0", entry.Seat)
	}
	assertNoUUID(t, entry.Text)
}

// A foretold card is face down in exile, and the printed cost names it
// as loudly as an `alternative_costs` entry would — which is why
// redactCardForViewer clears those. The line has to make the same trade.
func TestASpecialActionsLabelIsRedactedWithTheCard(t *testing.T) {
	g := buildActiveGame(t)
	actor, other := g.Seats[0], g.Seats[1]
	card := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{
			InstanceID: card, Name: "Saw It Coming", TypeLine: "Instant",
			Owner: actor.ID, Controller: actor.ID, FaceDown: true,
			KnownBy: map[uuid.UUID]bool{actor.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventSpecialAction, Actor: actor.ID, CardID: card, Label: "Foretell {2}",
		})
	})

	mine := findLog(t, FilterViewFor(ViewOfGame(g), actor.ID.String()).Log, LogSpecialAction)
	if want := "P1 used Foretell {2} on Saw It Coming"; mine.Text != want {
		t.Errorf("the actor's own line: got %q, want %q", mine.Text, want)
	}

	theirs := findLog(t, FilterViewFor(ViewOfGame(g), other.ID.String()).Log, LogSpecialAction)
	if want := "P1 took a special action on a card"; theirs.Text != want {
		t.Errorf("a non-knower's line: got %q, want %q", theirs.Text, want)
	}
	if theirs.Label != "" {
		t.Errorf("the printed cost survived the redaction as %q", theirs.Label)
	}
}

// --- 3. a cycling (CR 702.29b) ---------------------------------------

// The cost's discard already produced a zone line for the same motion.
// The cycling entry REPLACES it — one fact, one line, the way a
// sacrifice replaces the zone move it causes — so a Drake Haven table
// reads "cycled" rather than "was put into the graveyard".
func TestLogNarratesACyclingInPlaceOfItsDiscard(t *testing.T) {
	g := buildActiveGame(t)
	actor, other := g.Seats[0], g.Seats[1]
	card := uuid.New()
	g.WithWriteLock(func() {
		actor.Graveyard.PushTop(game.Card{
			InstanceID: card, Name: "Shefet Monitor", TypeLine: "Creature — Naga",
			Owner: actor.ID, Controller: actor.ID,
			KnownBy: map[uuid.UUID]bool{actor.ID: true, other.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventZoneMove, Actor: actor.ID, CardID: card,
			OldZone: game.ZoneHand, NewZone: game.ZoneGraveyard,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCycle, Actor: actor.ID, Source: card, CardID: card,
		})
	})

	log := ViewOfGame(g).Log
	entry := findLog(t, log, LogCycle)
	if want := "P1 cycled Shefet Monitor"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	for _, e := range log {
		if e.Kind == LogZone && e.CardID == card.String() {
			t.Errorf("the discard's zone line survived beside the cycling: %q", e.Text)
		}
	}
	assertNoUUID(t, entry.Text)
}

// --- 4. a counter landing --------------------------------------------

// The count on the wire is the one the engine's event carries: the
// total AFTER the change, because applyCounterLocked has no delta to
// give. The line says it the way a player reads the card.
func TestLogNarratesACounterLanding(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	bear := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", owner.ID)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventCounterPlaced, Target: bear, Label: game.CounterPlusOne, Amount: 2,
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogCounters)
	if want := "Grizzly Bears now has 2 +1/+1 counters"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Label != game.CounterPlusOne {
		t.Errorf("label: got %q, want the counter kind", entry.Label)
	}
	if entry.Amount != 2 {
		t.Errorf("amount: got %d, want 2 — the count after the change", entry.Amount)
	}
	if entry.CardID != bear.String() {
		t.Errorf("card_id: got %q, want the bear — the emitter names it on Target", entry.CardID)
	}
	if entry.Seat != NoSeat {
		t.Errorf("seat: got %d, want NoSeat — the event carries no actor", entry.Seat)
	}
	assertNoUUID(t, entry.Text)
}

// The last one coming off is the same line with the number the event
// carries, which is zero or less.
func TestLogNarratesTheLastCounterComingOff(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	bear := battlefieldCard(t, g, "Grizzly Bears", "Creature — Bear", owner.ID)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventCounterPlaced, Target: bear, Label: game.CounterMinusOne, Amount: 0,
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogCounters)
	if want := "Grizzly Bears has no -1/-1 counters left"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
}

// counterKindIsNarrated is the rule #1021 asked for before this arm
// could exist: loyalty and lore are already lines somewhere else, and
// a Commander turn moves dozens of both.
func TestLoyaltyAndLoreCountersAreNotLines(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	walker := battlefieldCard(t, g, "Test Walker", "Legendary Planeswalker — Test", owner.ID)
	saga := battlefieldCard(t, g, "Test Saga", "Enchantment — Saga", owner.ID)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventCounterPlaced, Target: walker, Label: game.CounterLoyalty, Amount: 4,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCounterPlaced, Target: saga, Label: game.CounterLore, Amount: 2,
		})
	})

	for _, e := range ViewOfGame(g).Log {
		if e.Kind == LogCounters {
			t.Errorf("a %s counter produced a line: %q", e.Label, e.Text)
		}
	}
}

// A counter on a card the viewer may not identify says neither the
// kind nor the count: redactCardForViewer strips CardView.counters from
// a non-knower, and a line that said them would hand back exactly what
// the card filter took away.
func TestACounterOnAnUnknownCardSaysNeitherKindNorCount(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0], g.Seats[1]
	hidden := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: hidden, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
			Owner: owner.ID, Controller: owner.ID, FaceDown: true,
			KnownBy: map[uuid.UUID]bool{owner.ID: true},
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCounterPlaced, Target: hidden, Label: game.CounterPlusOne, Amount: 3,
		})
	})

	mine := findLog(t, FilterViewFor(ViewOfGame(g), owner.ID.String()).Log, LogCounters)
	if want := "Grizzly Bears now has 3 +1/+1 counters"; mine.Text != want {
		t.Errorf("the knower's line: got %q, want %q", mine.Text, want)
	}

	theirs := findLog(t, FilterViewFor(ViewOfGame(g), other.ID.String()).Log, LogCounters)
	if want := "a card's counters changed"; theirs.Text != want {
		t.Errorf("a non-knower's line: got %q, want %q", theirs.Text, want)
	}
	if theirs.Label != "" || theirs.Amount != 0 {
		t.Errorf("the counter survived the redaction as %q x%d", theirs.Label, theirs.Amount)
	}
}

// --- 5. scry and surveil (CR 701.22, CR 701.25) -----------------------

// The cards are hidden at both ends and stay unnamed; the count the
// table watched move is public and is the whole entry.
func TestLogNarratesAScryWithoutNamingACard(t *testing.T) {
	g := buildActiveGame(t)
	actor := g.Seats[0]
	source := uuid.New()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventScry, Actor: actor.ID, Source: source, Amount: 1})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogScry)
	if want := "P1 scried and put 1 card on the bottom"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Amount != 1 {
		t.Errorf("amount: got %d, want 1", entry.Amount)
	}
	if entry.CardID != "" || entry.Target != "" {
		t.Errorf("a scry entry names a card: card_id %q target %q", entry.CardID, entry.Target)
	}
	assertNoUUID(t, entry.Text)
}

// A scry that kept everything on top still happened, and the decision
// is the half an opponent is entitled to read.
func TestLogNarratesAScryThatKeptEverything(t *testing.T) {
	g := buildActiveGame(t)
	actor := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventScry, Actor: actor.ID, Amount: 0})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogScry)
	if want := "P1 scried and kept every card on top"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
}

// Surveil is its own kind, for the reason game.EventSurveil is: two
// keywords, two payoffs, two different destinations for the cards that
// move.
func TestLogNarratesASurveilWithoutNamingACard(t *testing.T) {
	g := buildActiveGame(t)
	actor := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSurveil, Actor: actor.ID, Amount: 2})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogSurveil)
	if want := "P1 surveilled and put 2 cards into their graveyard"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.CardID != "" {
		t.Errorf("a surveil entry names a card: %q", entry.CardID)
	}
	assertNoUUID(t, entry.Text)
}

// --- 6. Saga chapters and Class levels --------------------------------

func TestLogNarratesASagaChapter(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	saga := battlefieldCard(t, g, "Urza's Saga", "Enchantment — Urza's Saga", owner.ID)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventSagaChapter, Actor: owner.ID,
			Source: saga, Target: saga, CardID: saga, Amount: 3,
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogSagaChapter)
	if want := "Urza's Saga reached chapter 3"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Amount != 3 {
		t.Errorf("amount: got %d, want the chapter number", entry.Amount)
	}
	if entry.Seat != 0 {
		t.Errorf("seat: got %d, want the Saga's controller", entry.Seat)
	}
	assertNoUUID(t, entry.Text)
}

func TestLogNarratesAClassLevel(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0]
	class := battlefieldCard(t, g, "Barbarian Class", "Enchantment — Class", owner.ID)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind: game.EventClassLevel, Actor: owner.ID,
			Source: class, Target: class, CardID: class, Amount: 2,
		})
	})

	entry := findLog(t, ViewOfGame(g).Log, LogClassLevel)
	if want := "Barbarian Class became level 2"; entry.Text != want {
		t.Errorf("text: got %q, want %q", entry.Text, want)
	}
	if entry.Amount != 2 {
		t.Errorf("amount: got %d, want the level", entry.Amount)
	}
	assertNoUUID(t, entry.Text)
}
