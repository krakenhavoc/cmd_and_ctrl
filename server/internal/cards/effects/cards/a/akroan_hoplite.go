package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akroan Hoplite — Creature — Human Soldier {R}{W}, 1/2 (Conspiracy: Take
// the Crown, S82 older sets batch 273, #5540):
//
//	"Whenever this creature attacks, it gets +X/+0 until end of turn, where X
//	 is the number of attacking creatures you control."
//
// X is counted as the trigger resolves (Kinscaer Sentry's count: every
// creature you control that is attacking, the Hoplite included), so a
// creature that has left combat no longer counts. The pump is pinned to the
// Hoplite as this object (CR 400.7).
//
// No simplification.
func init() {
	const label = "Akroan Hoplite — gets +X/+0 until end of turn, where X is the number of attacking creatures you control"
	effects.Register(effects.Spec{
		OracleID:     "e15db59f-4477-4139-8e7e-741ab4584e7f",
		Name:         "Akroan Hoplite",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks(label, func(g *game.Game, item *game.StackItem) error {
				x := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == item.Controller && c.IsCreature() && c.AttackingTarget != uuid.Nil {
						x++
					}
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: x, Label: label}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
