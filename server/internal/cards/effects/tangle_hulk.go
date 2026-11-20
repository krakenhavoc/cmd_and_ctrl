package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tangle Hulk — Artifact Creature — Phyrexian Beast {5}, 5/3 (Mirrodin Besieged,
// #5576):
//
//	"{2}{G}: Regenerate this creature."
//
// Undercity Troll's regeneration: a shield on the ability's own source
// (CR 701.19).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6d0360c-ebc6-48b5-80ad-d52dadcdd5f5",
		Name:         "Tangle Hulk",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{2}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
