package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infectious Host — Creature — Zombie {2}{B}, 1/1 (Ravnica: City of Guilds,
// #5600):
//
//	"When this creature dies, target player loses 2 life."
//
// A targeted dies trigger: a life loss, not damage, for the one target
// player still in the game as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "beb9c5c9-d4ab-423f-8937-91d308e09770",
		Name:         "Infectious Host",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Infectious Host — target player loses 2 life", effects.Do(effects.TargetPlayerLosesLife{N: 2})),
				effects.TargetPlayer("target player")),
		},
	})
}
