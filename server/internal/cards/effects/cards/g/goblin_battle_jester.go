package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "eba01ee1-55c4-48db-843b-fd6d48c32eda",
		Name:         "Goblin Battle Jester",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverYouCast(effects.OfColor("R"), label, effects.TargetCreatureCantBlockThisTurn(label)),
				effects.TargetCreature("target creature")),
		},
	})
}
