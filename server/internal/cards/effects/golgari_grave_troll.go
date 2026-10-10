package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golgari Grave-Troll — Creature — Troll Skeleton {4}{G}, 0/0:
//
//	"This creature enters with a +1/+1 counter on it for each creature
//	 card in your graveyard.
//	 {1}, Remove a +1/+1 counter from this creature: Regenerate this
//	 creature.
//	 Dredge 6 (If you would draw a card, you may mill six cards
//	 instead. If you do, return this card from your graveyard to your
//	 hand.)"
//
// Diregraf Colossus's counters (a CR 614 self-entry replacement counted
// as the Troll enters, so a creature milled in response is counted and
// a counter doubler sees the batch), a regeneration ability whose cost
// removes a counter at announce, and dredge 6. The Troll is a 0/0 that
// lives on its counters, so removing the last one is a real cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "686f4a37-b5e4-46d4-9e0e-11794b2d12cd",
		Name:         "Golgari Grave-Troll",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted("+1/+1", func(g *game.Game, src *game.Card) int {
				return len(gyCardIDs(g, src.Controller, func(c game.Card) bool { return c.IsCreature() }))
			}, "Golgari Grave-Troll: enters with a +1/+1 counter per creature card in your graveyard"),
			Dredge(6),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Remove a +1/+1 counter from this creature: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    Plus(ManaCost("{1}"), RemoveCountersFromThis("+1/+1", 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
