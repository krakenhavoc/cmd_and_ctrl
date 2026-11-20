package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phantom Beast — Creature — Illusion Beast {3}{U}, 4/5 (Magic 2011,
// #5579):
//
//	"When this creature becomes the target of a spell or ability,
//	 sacrifice it."
//
// Phantasmal Image's Illusion trigger, printed on the card: any spell or
// ability, its controller's own included, puts the trigger on the stack,
// and the Beast is sacrificed as it resolves if it is still there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13ca7c14-5db2-4d17-912a-fdeabcdcd9f8",
		Name:         "Phantom Beast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, Self, "Phantom Beast — sacrifice it", SacrificeThisIfStillOnBattlefield),
		},
	})
}
