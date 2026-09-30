package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// entry_discard_test.go — ADR 0098, the enumerator half. Mox Diamond's
// "you may discard a land card instead" (floor zero: the decline is
// AlwaysLegal) and Heart of Yavimaya's "sacrifice a Forest instead"
// (floor one: no unconditional answer, but its candidates cannot leave
// while the prompt blocks the table). Every offered answer is one the
// dispatcher accepts.

const (
	entryDiscardOracle   = "test-discard-rock"
	entrySacrificeOracle = "test-sacrifice-land"
)

func stubEntryChoiceCatalog(t *testing.T) {
	t.Helper()
	appliesToSelf := func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
		return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
			src != nil && ev.CardID == src.InstanceID
	}
	chooser := func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
		if ev != nil && ev.Actor != uuid.Nil {
			return ev.Actor
		}
		return src.Controller
	}
	toGraveyard := func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
		ev.NewZone = game.ZoneGraveyard
		ev.NewZoneOwner = src.Owner
		return nil
	}
	defs := map[string]*game.CardDef{
		entryDiscardOracle: {Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove}, SelfReplacement: true, Label: "Test Mox",
			EntryCardChoice: &game.EntryCardChoice{Action: game.EntryCardDiscard, Max: 1,
				Matches: func(c game.Card) bool { return c.IsLand() }},
			AppliesTo: appliesToSelf, Controller: chooser, Replace: toGraveyard,
		}}},
		entrySacrificeOracle: {Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove}, SelfReplacement: true, Label: "Test Heart",
			EntryCardChoice: &game.EntryCardChoice{Action: game.EntryCardSacrifice, Min: 1, Max: 1,
				Matches: func(c game.Card) bool { return c.IsLand() && c.HasSubtype("forest") }},
			AppliesTo: appliesToSelf, Controller: chooser, Replace: toGraveyard,
		}}},
	}
	prev := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef { return defs[key] }
	t.Cleanup(func() { game.CatalogLookup = prev })
}

func openEntryChoice(t *testing.T, g *game.Game, kind game.PendingChoiceKind, entering game.Card) *game.PendingChoice {
	t.Helper()
	me := g.Seats[0]
	id := handCard(me, entering)
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast / play: %v", err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	t.Fatalf("no %s prompt", kind)
	return nil
}

func TestEntryDiscardOffersTheDeclineAndEachLand(t *testing.T) {
	stubEntryChoiceCatalog(t)
	g := newTable(t)
	me := g.Seats[0]
	clearHand(me)
	handCard(me, basic("Forest", "Forest"))
	handCard(me, basic("Island", "Island"))
	handCard(me, creature("Bear", "{1}{G}", 2, 2))
	openEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand,
		game.Card{Name: "Test Mox", TypeLine: "Artifact", OracleID: entryDiscardOracle})

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	declines := 0
	for _, m := range moves {
		if len(pickedCardIDs(t, m)) == 0 {
			declines++
			if !m.AlwaysLegal {
				t.Errorf("the decline %q is not AlwaysLegal", m.Label)
			}
		}
	}
	if declines != 1 || len(moves) != 3 {
		t.Errorf("offered %v, want the decline and the two lands", labels(moves))
	}
}

func TestEntrySacrificeOffersEachForestAndNoDecline(t *testing.T) {
	stubEntryChoiceCatalog(t)
	g := newTable(t)
	me := g.Seats[0]
	clearHand(me)
	for _, name := range []string{"Forest A", "Forest B"} {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Basic Land — Forest",
			Owner: me.ID, Controller: me.ID})
	}
	openEntryChoice(t, g, game.PendingChoiceEntrySacrifice,
		game.Card{Name: "Test Heart", TypeLine: "Land", OracleID: entrySacrificeOracle})

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if len(moves) != 2 {
		t.Fatalf("offered %v, want one answer per Forest", labels(moves))
	}
	for _, m := range moves {
		if len(pickedCardIDs(t, m)) != 1 {
			t.Errorf("offered %q, which does not name exactly one Forest", m.Label)
		}
	}
}
