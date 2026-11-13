package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rabid Wolverines — Creature — Wolverine {3}{G}{G}, 4/4
// (Exodus, #5626):
//
//	"Whenever this creature becomes blocked by a creature, this creature
//	 gets +1/+1 until end of turn."
//
// Cave Tiger's trigger: per blocker (CR 509.3d), one EventBlock for each
// blocker–attacker pair, so blocked by two creatures it gets +2/+2. The
// Wolverines blocking is not them becoming blocked.
//
// No simplification.
func init() {
	const label = "Rabid Wolverines — +1/+1 until end of turn"
	Register(Spec{
		OracleID:     "ed076c39-9bfc-424a-8ce8-8572f7c0de73",
		Name:         "Rabid Wolverines",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBlock, viciousBattleragerBlocked, label, thisCreatureUntilEOT(label, 1, 1)),
		},
	})
}
