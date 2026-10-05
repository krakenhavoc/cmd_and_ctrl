package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// entry_devour_test.go — CR 702.82a, the enumerator half. Devour's
// prompt is "sacrifice any number of creatures": floor zero, so the
// decline is AlwaysLegal, and every offered subset is an answer the
// dispatcher accepts.

const entryDevourOracle = "test-devourer"

func TestEntryDevourOffersZeroAndEverySubset(t *testing.T) {
	appliesToSelf := func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
		return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
			src != nil && ev.CardID == src.InstanceID
	}
	defs := map[string]*game.CardDef{
		entryDevourOracle: {Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove}, SelfReplacement: true, Label: "Test Devourer",
			EntryCardChoice: &game.EntryCardChoice{Action: game.EntryCardSacrifice, AnyNumber: true,
				Matches: func(c game.Card) bool { return c.IsCreature() }},
			AppliesTo: appliesToSelf,
			Controller: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				if ev != nil && ev.Actor != uuid.Nil {
					return ev.Actor
				}
				return src.Controller
			},
		}}},
	}
	prev := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef { return defs[key] }
	t.Cleanup(func() { game.CatalogLookup = prev })

	g := newTable(t)
	me := g.Seats[0]
	clearHand(me)
	for _, name := range []string{"Goblin A", "Goblin B"} {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Goblin",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	c := openEntryChoice(t, g, game.PendingChoiceEntrySacrifice,
		game.Card{Name: "Test Devourer", TypeLine: "Creature — Hellion", OracleID: entryDevourOracle})
	if c.ChooseMin != 0 || c.ChooseMax != 2 {
		t.Errorf("bounds %d..%d, want 0..2", c.ChooseMin, c.ChooseMax)
	}

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	sizes := map[int]int{}
	for _, m := range moves {
		n := len(pickedCardIDs(t, m))
		sizes[n]++
		if n == 0 && !m.AlwaysLegal {
			t.Errorf("the zero answer %q is not AlwaysLegal", m.Label)
		}
	}
	if sizes[0] != 1 || sizes[1] != 2 || sizes[2] != 1 {
		t.Errorf("offered %v (sizes %v), want zero, each one, and both", labels(moves), sizes)
	}
}
