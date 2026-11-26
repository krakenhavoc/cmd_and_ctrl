package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goblin Locksmith — Creature — Goblin Rogue {1}{R}, 2/1 (Guilds of Ravnica,
// #5525):
//
//	"Whenever this creature attacks, creatures with defender can't block this
//	 turn."
//
// The #6237 seam: the filter is data, so the set stays live until cleanup
// (CR 611.2c) and a creature with defender that arrives this turn can't block
// either. It binds every player's creatures, the Locksmith's controller's too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d692adcf-8ab5-4208-85d1-c75285a85c59",
		Name:         "Goblin Locksmith",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Goblin Locksmith — creatures with defender can't block this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CantBlockThisTurnMatching(false, "Goblin Locksmith — creatures with defender can't block",
						game.PermanentQuery{Types: []string{"creature"}, Keyword: "defender"}).Apply(effects.NewContext(g, item))
				}),
		},
	})
}
