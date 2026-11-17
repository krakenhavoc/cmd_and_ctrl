package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragon Engine — Artifact Creature — Dragon Construct {3}, 1/3
// (Antiquities, #5640):
//
//	"{2}: This creature gets +1/+0 until end of turn."
//
// Hellkite Punisher's firebreathing for generic mana: each activation is
// its own +1/+0 until end of turn on the Engine, so they stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "727a6474-d5b5-42ab-ace3-352d28f499eb",
		Name:         "Dragon Engine",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Dragon Engine — +1/+0"}).Apply(NewContext(g, item))
			},
		}},
	})
}
