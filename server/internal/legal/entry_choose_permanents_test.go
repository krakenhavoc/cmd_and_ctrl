package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// entry_choose_permanents_test.go — ADR 0170, the enumerator half. An
// "any number" choice of other permanents as a permanent enters: floor
// zero, so the empty answer is AlwaysLegal, every offered set is one the
// dispatcher accepts, and the declaration's suggested set is offered
// first, whatever the subset walk reached.

const entryChooseOracle = "test-entry-choose"

// entryChooseSuggested is what the test suggestion recommends; set by the
// test before the prompt is queued.
var entryChooseSuggested uuid.UUID

func init() {
	game.RegisterEntryChoiceSuggestion(entryChooseOracle, func(*game.Game, uuid.UUID, []uuid.UUID) []uuid.UUID {
		return []uuid.UUID{entryChooseSuggested}
	})
}

func TestEntryChoosePermanentsOffersTheSuggestionFirst(t *testing.T) {
	defs := map[string]*game.CardDef{
		entryChooseOracle: {Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventZoneMove}, SelfReplacement: true, Label: "Test Chooser",
			EntryCardChoice: &game.EntryCardChoice{
				Action: game.EntryCardChoose, AnyNumber: true, AnyController: true,
				Matches:    func(c game.Card) bool { return c.IsCreature() },
				Then:       func(*game.Game, *game.ReplacementEvent, []uuid.UUID) error { return nil },
				SuggestKey: entryChooseOracle,
			},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield &&
					src != nil && ev.CardID == src.InstanceID
			},
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
	me, opp := g.Seats[0], g.Seats[1]
	clearHand(me)
	for _, name := range []string{"Goblin A", "Goblin B"} {
		g.Battlefield.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Goblin",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	theirs := uuid.New()
	entryChooseSuggested = theirs
	g.Battlefield.PushTop(game.Card{InstanceID: theirs, Name: "Their Goblin", TypeLine: "Creature — Goblin",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})

	c := openEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents,
		game.Card{Name: "Test Chooser", TypeLine: "Creature — Avatar", OracleID: entryChooseOracle})
	if c.ChooseMin != 0 || c.ChooseMax != 3 {
		t.Errorf("bounds %d..%d, want 0..3 (anyone's creatures)", c.ChooseMin, c.ChooseMax)
	}
	if len(c.ChooseSuggested) != 1 || c.ChooseSuggested[0] != theirs {
		t.Errorf("suggested %v, want the opponent's goblin", c.ChooseSuggested)
	}

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	var picks [][]uuid.UUID
	for _, m := range moves {
		ids := pickedCardIDs(t, m)
		if len(ids) == 0 && !m.AlwaysLegal {
			t.Errorf("the empty answer %q is not AlwaysLegal", m.Label)
		}
		picks = append(picks, ids)
	}
	if len(picks) != 8 {
		t.Errorf("offered %d answers (%v), want all 8 subsets", len(picks), labels(moves))
	}
	if len(picks) < 2 || len(picks[1]) != 1 || picks[1][0] != theirs {
		t.Errorf("the suggested set is not the first non-empty answer: %v", labels(moves))
	}
}
