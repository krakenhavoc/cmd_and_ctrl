package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Battle Jester — Creature — Goblin {3}{R}, 2/2 (Magic 2013, #5568):
//
//	"Whenever you cast a red spell, target creature can't block this turn."
//
// Balefire Liege's red-spell trigger with Ahn-Crop Crasher's body: the
// target, if still legal as the trigger resolves, can't block until end
// of turn.
//
// No simplification.
func init() {
	const label = "Goblin Battle Jester — target creature can't block this turn"
	Register(Spec{
		OracleID:     "eba01ee1-55c4-48db-843b-fd6d48c32eda",
		Name:         "Goblin Battle Jester",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverYouCast(OfColor("R"), label, targetCreatureCantBlockThisTurn(label)),
				TargetCreature("target creature")),
		},
	})
}
