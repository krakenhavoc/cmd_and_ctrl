package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "69c22788-cb90-4100-a64d-891124082dd4",
		Name:            "Ruthless Deathfang",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventSacrifice, jmpB1YouSacrificedACreature, ruthlessDeathfangLabel,
				targetPlayerSacrificesACreature(ruthlessDeathfangLabel)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
