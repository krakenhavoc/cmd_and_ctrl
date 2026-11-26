package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruinous Minotaur — Creature — Minotaur Warrior {1}{R}{R} 5/2 (Zendikar, S82
// OLD batch 492, #5583):
//
//	"Whenever this creature deals damage to an opponent, sacrifice a land."
//
// One trigger per damage event to an opponent (combat or not); its controller
// chooses which land to sacrifice as it resolves, and nothing happens with no
// land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ece9d3e5-79e5-4c56-9e8d-8abcd819c041",
		Name:         "Ruinous Minotaur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsDamageToAnOpponent("Ruinous Minotaur — sacrifice a land",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					lands := effects.LandsControlledByPlayer(g, ctx.Controller())
					if len(lands) == 0 {
						return nil
					}
					return effects.SacrificeChoice{
						Player:     ctx.Controller(),
						Candidates: lands,
						Question:   "Ruinous Minotaur — sacrifice a land",
					}.Apply(ctx)
				}),
		},
	})
}
