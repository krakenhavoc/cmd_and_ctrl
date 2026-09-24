package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_lands_test.go — "as this land enters, you may reveal an
// Island or Swamp card from your hand. If you don't, it enters
// tapped."
//
// The assertions that earn their keep are shocklands_test.go's, for
// its reason: the tap EVENT count, not the Tapped flag, is what tells
// a replaced entry from an enter-tapped-then-untap patch-up, and the
// prompt pausing the entry — nothing on the battlefield, the card
// still in hand — is what tells a CR 614 replacement from an ETB
// trigger. Two more belong to this seam alone: the revealed card is
// still in the revealer's HAND afterwards (a reveal is not a zone
// change, CR 701.20b), and the table can read it (CR 701.20).

// revealLandOracleIDs is the whole cycle, name → oracle ID, verified
// against Scryfall.
var revealLandOracleIDs = map[string]string{
	"Choked Estuary":     "d473b507-8c33-4118-bc10-b0a268776074",
	"Foreboding Ruins":   "5c87e2fa-77f1-4978-b25f-f14d227301d1",
	"Port Town":          "458d2b12-f578-4392-98d3-c3bc83f316c4",
	"Fortified Village":  "56f1a16a-9f41-41fb-b580-c200bca27cd6",
	"Frostboil Snarl":    "7137aae6-260d-41de-8b4e-42a8cf752697",
	"Furycalm Snarl":     "651dea9c-2375-4e44-8e65-ba8e40f0c0ef",
	"Game Trail":         "00de57d2-7cb6-4337-9bc6-f6711e4dfabf",
	"Shineshadow Snarl":  "c9fc13d6-bd10-47bc-b2b6-7f67a1f3371e",
	"Necroblossom Snarl": "761ee6f9-b0fa-43c9-8d1f-9591ea18e52d",
	"Vineglimmer Snarl":  "33f52df8-4b44-4422-8b0a-37fead9c894b",
}

// entryRevealChoiceFor returns the open "you may reveal …" prompt
// addressed to chooser, or nil.
func entryRevealChoiceFor(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for i := len(g.PendingChoices) - 1; i >= 0; i-- {
		c := g.PendingChoices[i]
		if c != nil && c.Kind == game.PendingChoiceEntryRevealFromHand && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

// answerEntryReveal answers the open entry prompt for chooser. An
// empty `picks` is the decline.
func answerEntryReveal(t *testing.T, g *game.Game, chooser uuid.UUID, picks ...uuid.UUID) {
	t.Helper()
	c := entryRevealChoiceFor(g, chooser)
	if c == nil {
		t.Fatalf("no PendingChoiceEntryRevealFromHand addressed to %s", chooser)
	}
	if err := g.ResolveEntryRevealFromHand(c.ID, chooser, picks); err != nil {
		t.Fatalf("ResolveEntryRevealFromHand: %v", err)
	}
}

// revealEventsFor counts EventRevealCards entries for one card.
func revealEventsFor(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards && ev.CardID == id {
			n++
		}
	}
	return n
}

// findCardInHand returns a pointer to the live card in a player's
// hand, so a test can read the knower set the reveal wrote.
func findCardInHand(p *game.Player, id uuid.UUID) *game.Card {
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			return &p.Hand.Cards[i]
		}
	}
	return nil
}

// seedRevealLandHand puts one card of each named type line into the
// active seat's hand and returns the IDs in the order given.
func seedRevealLandHand(p *game.Player, typeLines ...string) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(typeLines))
	for i, tl := range typeLines {
		out = append(out, handCard(p, "fixture "+string(rune('a'+i)), tl))
	}
	return out
}

func TestRevealLandPromptPausesTheEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	island := seedRevealLandHand(me, "Basic Land — Island")[0]

	id := playLandFromHand(t, g, "Choked Estuary", revealLandOracleIDs["Choked Estuary"])

	c := entryRevealChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("playing a reveal-land queued no reveal prompt")
	}
	if c.Source != id {
		t.Errorf("prompt source: got %s, want the entering land %s", c.Source, id)
	}
	if c.ChooseMin != 0 {
		t.Errorf("prompt floor: got %d, want 0 — declining is a real answer", c.ChooseMin)
	}
	if c.ChooseMax != 1 {
		t.Errorf("prompt ceiling: got %d, want 1", c.ChooseMax)
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != island {
		t.Errorf("candidates: got %v, want just the Island %s", c.ChooseCards, island)
	}
	// The entry is what's being replaced, so nothing has entered.
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("the land reached the battlefield before the choice was made")
	}
	if !me.Hand.Contains(id) {
		t.Error("the land left the hand while the entry prompt was still open")
	}
	// And the table is stopped until it is answered (#791's gate).
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("PassPriority with the entry prompt open: got %v, want ErrChoicePending", err)
	}

	answerEntryReveal(t, g, me.ID)
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("answering the prompt did not resume the entry")
	}
}

func TestRevealLandRevealedEntersUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := seedRevealLandHand(me, "Basic Land — Swamp")[0]

	id := playLandFromHand(t, g, "Foreboding Ruins", revealLandOracleIDs["Foreboding Ruins"])
	answerEntryReveal(t, g, me.ID, swamp)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Foreboding Ruins is not on the battlefield")
	}
	if card.Tapped {
		t.Error("revealed a Swamp and the land still entered tapped")
	}
	// The discriminator against an enter-tapped-then-untap patch-up:
	// a replaced entry taps and untaps nothing at all.
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; a revealed reveal-land is never tapped at all", n)
	}
	if n := untapEventsFor(g, id); n != 0 {
		t.Errorf("%d untap events; the land entered untapped", n)
	}
	// A reveal is not a zone change (CR 701.20b).
	if !me.Hand.Contains(swamp) {
		t.Error("the revealed card left the hand; revealing moves nothing")
	}
}

// TestRevealLandNarratesTheReveal is the other-seats half of the
// prompt's per-viewer split: the picker is the revealer's alone, so
// the table has to learn what was shown some other way. It does, from
// the log, through the one reveal primitive — which also makes every
// seat a knower (CR 701.20, "entitled to remember").
func TestRevealLandNarratesTheReveal(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := seedRevealLandHand(me, "Basic Land — Plains")[0]

	playLandFromHand(t, g, "Port Town", revealLandOracleIDs["Port Town"])
	answerEntryReveal(t, g, me.ID, plains)

	if n := revealEventsFor(g, plains); n != 1 {
		t.Errorf("%d reveal events for the named card; want exactly 1", n)
	}
	card := findCardInHand(me, plains)
	if card == nil {
		t.Fatal("the revealed card is gone from the hand")
	}
	for _, p := range g.Seats {
		if !card.IsKnownTo(p.ID) {
			t.Errorf("seat %s is not a knower of the revealed card", p.Name)
		}
	}
}

func TestRevealLandDeclinedEntersTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := seedRevealLandHand(me, "Basic Land — Forest")[0]

	id := playLandFromHand(t, g, "Fortified Village", revealLandOracleIDs["Fortified Village"])
	answerEntryReveal(t, g, me.ID)

	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("Fortified Village is not on the battlefield")
	}
	if !card.Tapped {
		t.Error("declined the reveal and the land entered untapped anyway")
	}
	// The discriminator: the land ENTERED tapped, so nothing tapped
	// it. An OnETB tap would show up here as one event.
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
	// Declining is a real answer, not a forced reveal: the Forest is
	// still unknown to everyone else.
	if n := revealEventsFor(g, forest); n != 0 {
		t.Errorf("%d reveal events on a decline; want 0", n)
	}
}

// TestRevealLandWithNothingMatchingIsNotPrompted is the empty-hand
// case, and it is the ABSENCE of the question rather than a refusal
// of it: asking a question whose only answer is "no" is worse than
// not asking, so the land just enters tapped.
func TestRevealLandWithNothingMatchingIsNotPrompted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	seedRevealLandHand(me, "Basic Land — Forest", "Creature — Goblin")

	id := playLandFromHand(t, g, "Frostboil Snarl", revealLandOracleIDs["Frostboil Snarl"])

	if c := entryRevealChoiceFor(g, me.ID); c != nil {
		t.Error("prompted with nothing in hand that the clause admits")
	}
	card, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the land did not enter when the prompt was skipped")
	}
	if !card.Tapped {
		t.Error("un-revealed reveal-land entered untapped")
	}
	if n := tapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
}

// TestRevealLandFilterOffersOnlyWhatTheClauseAdmits is the filter
// itself. Game Trail names a Mountain or Forest card; a Plains, an
// Island and a Goblin are not candidates, and neither is a MOUNTAIN
// GIANT — the clause is a land-type test, not a name match.
func TestRevealLandFilterOffersOnlyWhatTheClauseAdmits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	ids := seedRevealLandHand(me,
		"Basic Land — Mountain",
		"Land — Forest Island",
		"Basic Land — Plains",
		"Creature — Goblin",
	)
	mountain, forestIsland := ids[0], ids[1]
	giant := handCard(me, "Mountain Giant", "Creature — Giant")

	playLandFromHand(t, g, "Game Trail", revealLandOracleIDs["Game Trail"])

	c := entryRevealChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no reveal prompt")
	}
	want := map[uuid.UUID]bool{mountain: true, forestIsland: true}
	if len(c.ChooseCards) != len(want) {
		t.Fatalf("candidates: got %d (%v), want %d", len(c.ChooseCards), c.ChooseCards, len(want))
	}
	for _, id := range c.ChooseCards {
		if !want[id] {
			t.Errorf("candidate %s is not a Mountain or Forest card", id)
		}
	}
	// And the resolver refuses what the prompt did not offer, rather
	// than trusting the client's list.
	if err := g.ResolveEntryRevealFromHand(c.ID, me.ID, []uuid.UUID{giant}); err == nil {
		t.Error("revealing a card the clause does not admit was accepted")
	}
	if entryRevealChoiceFor(g, me.ID) == nil {
		t.Error("a refused answer lost the prompt; it must stay open for another try")
	}
}

// TestRevealLandNoStackTrip pins the CR 614 half: the choice is not a
// triggered ability, so it never touches the stack and never hands an
// opponent priority.
func TestRevealLandNoStackTrip(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountain := seedRevealLandHand(me, "Basic Land — Mountain")[0]

	id := playLandFromHand(t, g, "Furycalm Snarl", revealLandOracleIDs["Furycalm Snarl"])
	answerEntryReveal(t, g, me.ID, mountain)

	if n := g.Stack.Size(); n != 0 {
		t.Errorf("stack holds %d items; the entry choice is a replacement, not a trigger", n)
	}
	if len(g.PendingTriggers) != 0 {
		t.Errorf("%d pending triggers; the entry choice is not a trigger", len(g.PendingTriggers))
	}
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("Furycalm Snarl is not on the battlefield")
	}
}

// TestRevealLandUndoAcrossThePrompt: the whole window is rewindable.
// A clone taken while the question is open, restored after it was
// answered, puts the land back in hand with the question still owed
// and the revealed card unknown to the table again.
func TestRevealLandUndoAcrossThePrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := seedRevealLandHand(me, "Basic Land — Swamp")[0]
	opponent := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	id := playLandFromHand(t, g, "Shineshadow Snarl", revealLandOracleIDs["Shineshadow Snarl"])
	if entryRevealChoiceFor(g, me.ID) == nil {
		t.Fatal("no reveal prompt to snapshot across")
	}
	snap := g.Clone()

	answerEntryReveal(t, g, me.ID, swamp)
	if card, ok := battlefieldCard(g, id); !ok || card.Tapped {
		t.Fatalf("setup: the land should be on the battlefield untapped (ok=%v)", ok)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	me = g.Seats[g.Turn.ActiveSeat]
	if entryRevealChoiceFor(g, me.ID) == nil {
		t.Error("undo did not restore the open reveal prompt")
	}
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("undo left the land on the battlefield")
	}
	if !me.Hand.Contains(id) {
		t.Error("undo did not put the land back in hand")
	}
	if card := findCardInHand(me, swamp); card == nil {
		t.Error("undo lost the revealed card")
	} else if card.IsKnownTo(opponent.ID) {
		t.Error("undo did not roll back the reveal's knowledge")
	}
}

// TestRevealLandDepartureSettlesTheEntry is the CR 800.4a case. The
// prompt is dropped rather than reassigned — the hand and the land
// are both the departed player's — and the paused entry is settled by
// the replacement frame the prompt carries, not left hanging.
func TestRevealLandDepartureSettlesTheEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedRevealLandHand(me, "Basic Land — Swamp")

	id := playLandFromHand(t, g, "Necroblossom Snarl", revealLandOracleIDs["Necroblossom Snarl"])
	if entryRevealChoiceFor(g, me.ID) == nil {
		t.Fatal("no reveal prompt")
	}

	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if c := entryRevealChoiceFor(g, me.ID); c != nil {
		t.Error("the prompt survived its chooser leaving the game")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceEntryRevealFromHand {
			t.Errorf("the prompt was reassigned to %s; nobody else can answer it", c.Chooser)
		}
	}
	// The land belonged to the seat that left, so CR 800.4a takes it
	// out of the game with them. What matters is that the paused
	// event was SETTLED — nothing is left blocking the table.
	if _, ok := battlefieldCard(g, id); ok {
		t.Error("the land stayed on the battlefield after its owner left (CR 800.4a)")
	}
	if err := g.PassPriority(); errors.Is(err, game.ErrChoicePending) {
		t.Error("a prompt is still blocking the table after its chooser left")
	}
}

// TestEveryRevealLandIsRegistered holds the cycle against the
// catalog: ten cards, one clause, and — since #1322 made the "put onto
// the battlefield" batch ask its question — each of them complete.
func TestEveryRevealLandIsRegistered(t *testing.T) {
	for name, oracleID := range revealLandOracleIDs {
		spec, ok := Lookup(oracleID)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracleID)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s: registered under the name %q", name, spec.Name)
		}
		if len(spec.Replacements) != 1 || spec.Replacements[0].EntryHandReveal == nil {
			t.Errorf("%s: no EntryHandReveal clause", name)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %q with %d caveats", name, spec.Completeness, len(spec.Caveats))
		}
	}
}
