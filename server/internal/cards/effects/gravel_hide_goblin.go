package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gravel-Hide Goblin — Creature — Goblin Shaman {1}{R}, 2/1 (Ravnica:
// Clue Edition, #5929):
//
//	"{3}{G}: This creature gets +2/+2 until end of turn."
//
// Shivan Dragon's untapped pump at a different price and size.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71d284ee-5b29-4db3-b355-a6bdc142a05b",
		Name:         "Gravel-Hide Goblin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}: This creature gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2, Toughness: 2}},
			Cost:    ManaCost("{3}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Gravel-Hide Goblin — +2/+2"}.Apply(ctx)
			},
		}},
	})
}
