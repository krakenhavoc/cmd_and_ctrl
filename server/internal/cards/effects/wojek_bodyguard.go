package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wojek Bodyguard — Creature — Human Soldier {2}{R}, 3/3 (Ravnica: Clue
// Edition):
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on target
//	 attacking creature with lesser power.)
//	 This creature can't attack or block alone."
//
// Barging Sergeant's mentor trigger with Pet Sehlat's attack limit and Toby's
// Beast's block rule (CR 506.5, 508.1c, 509.1b): two creatures that each
// can't attack or block alone may attack or block together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8de57b66-cb97-44e4-8c0d-8ecf3202c299",
		Name:         "Wojek Bodyguard",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Mentor("Wojek Bodyguard — mentor")},
		AttackLimits: []game.AttackLimit{CantAttackAlone()},
		BlockRules:   []game.BlockRule{CantBlockAlone()},
	})
}
