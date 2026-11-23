package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const ruthlessDeathfangLabel = "Ruthless Deathfang — target opponent sacrifices a creature"

// Ruthless Deathfang — Creature — Dragon {4}{U}{B}, 4/4:
//
//	"Flying
//	 Whenever you sacrifice a creature, target opponent sacrifices a
//	 creature of their choice."
//
// One trigger per creature you sacrifice (a multi-creature cost is one
// event each). The sacrifice event fires while the creature is still
// on the battlefield, so the Deathfang sacrificing itself still
// triggers. The opponent chooses what to sacrifice as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "69c22788-cb90-4100-a64d-891124082dd4",
		Name:            "Ruthless Deathfang",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventSacrifice, effects.JmpB1YouSacrificedACreature, ruthlessDeathfangLabel,
				effects.TargetPlayerSacrificesACreature(ruthlessDeathfangLabel)),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
