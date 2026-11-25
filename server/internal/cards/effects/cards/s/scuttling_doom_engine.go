package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scuttling Doom Engine — Artifact Creature — Construct {6}, 6/6:
//
//	"This creature can't be blocked by creatures with power 2 or less.
//	 When this creature dies, it deals 6 damage to target opponent or
//	 planeswalker."
//
// The block restriction is Questing Beast's shared CantBeBlockedBy. The dies
// trigger names its target as it goes on the stack (CR 603.3d) and deals the
// damage as the creature it was (CR 608.2h); a target that has left is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72ccdcef-5c84-45ba-bda3-3dc014e1585c",
		Name:         "Scuttling Doom Engine",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(2), "creatures with power 2 or less")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Scuttling Doom Engine — it deals 6 damage to target opponent or planeswalker",
				effects.DealDamageToTheTarget(6)), effects.TargetOpponentOrPlaneswalker()),
		},
	})
}
