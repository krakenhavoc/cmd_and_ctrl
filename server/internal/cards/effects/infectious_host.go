package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "beb9c5c9-d4ab-423f-8937-91d308e09770",
		Name:         "Infectious Host",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Infectious Host — target player loses 2 life", Do(TargetPlayerLosesLife{N: 2})),
				TargetPlayer("target player")),
		},
	})
}
