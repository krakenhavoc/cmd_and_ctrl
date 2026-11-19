package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slimy Piper — Creature — Fungus Bard {1}{G}, 2/1 (Foundations Jumpstart):
//
//	"Whenever this creature attacks, it gets +1/+1 until end of turn. If
//	 you control four or more creatures, it gets +2/+2 and gains
//	 indestructible until end of turn instead."
//
// One trigger; the creature count is read as it resolves (it is not an
// intervening "if"), so a creature that died or arrived in response counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be0c9947-3e7c-4b53-b87c-ff16ffd668ea",
		Name:         "Slimy Piper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Slimy Piper — +1/+1 until end of turn (+2/+2 and indestructible with four or more creatures)", slimyPiperAttack),
		},
	})
}
