package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pygmy Troll — Creature — Troll {1}{G}, 1/1 (Exodus, #5626):
//
//	"Whenever this creature becomes blocked by a creature, this creature
//	 gets +1/+1 until end of turn.
//	 {G}: Regenerate this creature."
//
// Cave Tiger's per-blocker trigger (CR 509.3d) and the shared
// regenerateThisRow (CR 701.19a).
//
// No simplification.
func init() {
	const label = "Pygmy Troll — +1/+1 until end of turn"
	Register(Spec{
		OracleID:     "6f18f41e-df8d-4b59-87cc-ea167823fbda",
		Name:         "Pygmy Troll",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, viciousBattleragerBlocked, label, thisCreatureUntilEOT(label, 1, 1)),
		},
		Activated: []ActivatedAbility{regenerateThisRow("{G}")},
	})
}
