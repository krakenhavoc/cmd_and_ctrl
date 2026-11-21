package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauntless Aven — Creature — Bird Warrior {2}{W}, 2/1 (Hour of
// Devastation, #5534):
//
//	"Flying
//	 Whenever this creature attacks, untap target creature you control."
//
// Flying is a printed keyword. The attack trigger is Civic Gardener's with
// a creature-you-control clause; the Aven may target itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7104e471-06aa-426d-af1e-d5f0b1fb18c3",
		Name:            "Dauntless Aven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Dauntless Aven — untap target creature you control", effects.UntapTheTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
