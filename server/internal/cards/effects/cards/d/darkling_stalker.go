package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darkling Stalker — Creature — Shade Spirit {3}{B}, 1/1 (Tempest, S82 OLD
// b755, #5628):
//
//	"{B}: Regenerate this creature.
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Screeching Harpy's regeneration and the shared self pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0103b2f2-05dd-40af-954d-303c144edb04",
		Name:         "Darkling Stalker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{B}: Regenerate this creature.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.RegenerateThisSource,
			},
			{
				Label:   "{B}: This creature gets +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
				Cost:    effects.ManaCost("{B}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Darkling Stalker — +1/+1 until end of turn"),
			},
		},
	})
}
