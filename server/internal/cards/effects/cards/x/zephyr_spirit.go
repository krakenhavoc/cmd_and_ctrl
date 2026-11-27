package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephyr Spirit — Creature — Spirit {5}{U}, 0/6 (Ravnica: City of Guilds):
//
//	"When this creature blocks, return it to its owner's hand."
//
// One trigger per block declaration (a creature that blocks several attackers
// blocks once, CR 509.3a). It resolves with the Spirit still blocking, then the
// bounce takes it out of combat; the attacker stays blocked (CR 509.1h). A Spirit
// that has already left, or left and came back as a new object, is not returned
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ff640343-6980-4668-a41f-0f01b0bbc0f2",
		Name:         "Zephyr Spirit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Zephyr Spirit — return it to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BounceToHand{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
