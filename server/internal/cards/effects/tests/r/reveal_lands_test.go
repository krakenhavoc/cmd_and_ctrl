package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

func TestRevealLandPromptPausesTheEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	island := effectstest.SeedRevealLandHand(me, "Basic Land — Island")[0]

	id := effectstest.PlayLandFromHand(t, g, "Choked Estuary", effectstest.RevealLandOracleIDs["Choked Estuary"])

	c := effectstest.EntryRevealChoiceFor(g, me.ID)
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
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
		t.Error("the land reached the battlefield before the choice was made")
	}
	if !me.Hand.Contains(id) {
		t.Error("the land left the hand while the entry prompt was still open")
	}
	// And the table is stopped until it is answered (#791's gate).
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("PassPriority with the entry prompt open: got %v, want ErrChoicePending", err)
	}

	effectstest.AnswerEntryReveal(t, g, me.ID)
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Fatal("answering the prompt did not resume the entry")
	}
}

func TestRevealLandRevealedEntersUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := effectstest.SeedRevealLandHand(me, "Basic Land — Swamp")[0]

	id := effectstest.PlayLandFromHand(t, g, "Foreboding Ruins", effectstest.RevealLandOracleIDs["Foreboding Ruins"])
	effectstest.AnswerEntryReveal(t, g, me.ID, swamp)

	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("Foreboding Ruins is not on the battlefield")
	}
	if card.Tapped {
		t.Error("revealed a Swamp and the land still entered tapped")
	}
	// The discriminator against an enter-tapped-then-untap patch-up:
	// a replaced entry taps and untaps nothing at all.
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; a revealed reveal-land is never tapped at all", n)
	}
	if n := effectstest.UntapEventsFor(g, id); n != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plains := effectstest.SeedRevealLandHand(me, "Basic Land — Plains")[0]

	effectstest.PlayLandFromHand(t, g, "Port Town", effectstest.RevealLandOracleIDs["Port Town"])
	effectstest.AnswerEntryReveal(t, g, me.ID, plains)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := effectstest.SeedRevealLandHand(me, "Basic Land — Forest")[0]

	id := effectstest.PlayLandFromHand(t, g, "Fortified Village", effectstest.RevealLandOracleIDs["Fortified Village"])
	effectstest.AnswerEntryReveal(t, g, me.ID)

	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("Fortified Village is not on the battlefield")
	}
	if !card.Tapped {
		t.Error("declined the reveal and the land entered untapped anyway")
	}
	// The discriminator: the land ENTERED tapped, so nothing tapped
	// it. An OnETB tap would show up here as one event.
	if n := effectstest.TapEventsFor(g, id); n != 0 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	effectstest.SeedRevealLandHand(me, "Basic Land — Forest", "Creature — Goblin")

	id := effectstest.PlayLandFromHand(t, g, "Frostboil Snarl", effectstest.RevealLandOracleIDs["Frostboil Snarl"])

	if c := effectstest.EntryRevealChoiceFor(g, me.ID); c != nil {
		t.Error("prompted with nothing in hand that the clause admits")
	}
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("the land did not enter when the prompt was skipped")
	}
	if !card.Tapped {
		t.Error("un-revealed reveal-land entered untapped")
	}
	if n := effectstest.TapEventsFor(g, id); n != 0 {
		t.Errorf("%d tap events; the land should have ENTERED tapped", n)
	}
}

// TestRevealLandFilterOffersOnlyWhatTheClauseAdmits is the filter
// itself. Game Trail names a Mountain or Forest card; a Plains, an
// Island and a Goblin are not candidates, and neither is a MOUNTAIN
// GIANT — the clause is a land-type test, not a name match.
func TestRevealLandFilterOffersOnlyWhatTheClauseAdmits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.Cards = nil
	ids := effectstest.SeedRevealLandHand(me,
		"Basic Land — Mountain",
		"Land — Forest Island",
		"Basic Land — Plains",
		"Creature — Goblin",
	)
	mountain, forestIsland := ids[0], ids[1]
	giant := effectstest.HandCard(me, "Mountain Giant", "Creature — Giant")

	effectstest.PlayLandFromHand(t, g, "Game Trail", effectstest.RevealLandOracleIDs["Game Trail"])

	c := effectstest.EntryRevealChoiceFor(g, me.ID)
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
	if effectstest.EntryRevealChoiceFor(g, me.ID) == nil {
		t.Error("a refused answer lost the prompt; it must stay open for another try")
	}
}

// TestRevealLandNoStackTrip pins the CR 614 half: the choice is not a
// triggered ability, so it never touches the stack and never hands an
// opponent priority.
func TestRevealLandNoStackTrip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mountain := effectstest.SeedRevealLandHand(me, "Basic Land — Mountain")[0]

	id := effectstest.PlayLandFromHand(t, g, "Furycalm Snarl", effectstest.RevealLandOracleIDs["Furycalm Snarl"])
	effectstest.AnswerEntryReveal(t, g, me.ID, mountain)

	if n := g.Stack.Size(); n != 0 {
		t.Errorf("stack holds %d items; the entry choice is a replacement, not a trigger", n)
	}
	if len(g.PendingTriggers) != 0 {
		t.Errorf("%d pending triggers; the entry choice is not a trigger", len(g.PendingTriggers))
	}
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Fatal("Furycalm Snarl is not on the battlefield")
	}
}

// TestRevealLandUndoAcrossThePrompt: the whole window is rewindable.
// A clone taken while the question is open, restored after it was
// answered, puts the land back in hand with the question still owed
// and the revealed card unknown to the table again.
func TestRevealLandUndoAcrossThePrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swamp := effectstest.SeedRevealLandHand(me, "Basic Land — Swamp")[0]
	opponent := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	id := effectstest.PlayLandFromHand(t, g, "Shineshadow Snarl", effectstest.RevealLandOracleIDs["Shineshadow Snarl"])
	if effectstest.EntryRevealChoiceFor(g, me.ID) == nil {
		t.Fatal("no reveal prompt to snapshot across")
	}
	snap := g.Clone()

	effectstest.AnswerEntryReveal(t, g, me.ID, swamp)
	if card, ok := effectstest.BattlefieldCard(g, id); !ok || card.Tapped {
		t.Fatalf("setup: the land should be on the battlefield untapped (ok=%v)", ok)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })

	me = g.Seats[g.Turn.ActiveSeat]
	if effectstest.EntryRevealChoiceFor(g, me.ID) == nil {
		t.Error("undo did not restore the open reveal prompt")
	}
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedRevealLandHand(me, "Basic Land — Swamp")

	id := effectstest.PlayLandFromHand(t, g, "Necroblossom Snarl", effectstest.RevealLandOracleIDs["Necroblossom Snarl"])
	if effectstest.EntryRevealChoiceFor(g, me.ID) == nil {
		t.Fatal("no reveal prompt")
	}

	if err := g.Concede(me.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}

	if c := effectstest.EntryRevealChoiceFor(g, me.ID); c != nil {
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
	if _, ok := effectstest.BattlefieldCard(g, id); ok {
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
	for name, oracleID := range effectstest.RevealLandOracleIDs {
		spec, ok := effects.Lookup(oracleID)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracleID)
			continue
		}
		if spec.Name != name {
			t.Errorf("%s: registered under the name %q", name, spec.Name)
		}
		if len(spec.Replacements) != 1 || spec.Replacements[0].EntryCardChoice == nil {
			t.Errorf("%s: no EntryCardChoice clause", name)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %q with %d caveats", name, spec.Completeness, len(spec.Caveats))
		}
	}
}
