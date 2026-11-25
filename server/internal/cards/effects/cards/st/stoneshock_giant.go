package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stoneshock Giant — Creature — Giant {3}{R}{R}, 5/4 (Conspiracy: Take the
// Crown, S82 older sets batch 271, #5540):
//
//	"{6}{R}{R}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)
//	 When this creature becomes monstrous, creatures without flying your
//	 opponents control can't block this turn."
//
// The shared Monstrosity activation and WhenBecomesMonstrous (CR 701.37c).
// The body is Rumbling Ruin's can't-block over a live query of the
// opponents' creatures, here "without flying" (PermanentQuery.WithoutKeyword):
// the set stays live until cleanup (CR 509.1b), so a creature without flying
// that arrives later this turn can't block either, and a creature that
// gains flying this turn can.
//
// No simplification.
func init() {
	const label = "Stoneshock Giant — creatures without flying your opponents control can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "dd2c0046-dc31-44f7-8601-7ec77515b926",
		Name:         "Stoneshock Giant",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{6}{R}{R}"), 3)},
		Triggered: []game.TriggeredAbility{
			effects.WhenBecomesMonstrous(label, func(g *game.Game, item *game.StackItem) error {
				return effects.CantBlockThisTurnMatching(true, label,
					game.PermanentQuery{Types: []string{"creature"}, WithoutKeyword: "flying"}).Apply(effects.NewContext(g, item))
			}),
		},
	})
}
