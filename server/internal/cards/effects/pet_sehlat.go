package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pet Sehlat — Creature — Bear {2}{W}, 4/3:
//
//	"This creature can't attack alone."
//
// #4973. A restriction on the creature (CR 508.1c, 506.5): the attack
// limit CantAttackAlone (game.AttackLimitNotAlone). A declaration whose
// only attacker is the Sehlat is refused at its checkpoint while another
// creature could still join it, the legal-move enumerator then offers
// only those companions, and the verbs never begin a declaration with the
// Sehlat when nothing could join it. A CR 508.1d requirement on it counts
// only where a companion can attack beside it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "509f255f-c3ee-4836-a583-d018060d0c4d",
		Name:         "Pet Sehlat",
		Completeness: CompletenessFull,
		AttackLimits: []game.AttackLimit{CantAttackAlone()},
	})
}
