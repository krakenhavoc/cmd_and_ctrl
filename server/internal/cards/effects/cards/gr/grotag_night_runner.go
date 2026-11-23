package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grotag Night-Runner — Creature — Goblin Rogue {2}{R}, 2/3 (Jumpstart 2022,
// #5948):
//
//	"Whenever this creature deals combat damage to a player, exile the top
//	 card of your library. You may play that card this turn."
//
// Bell Borca's impulse draw on a combat-damage trigger: one trigger per combat
// damage event, the top card exiled face up and playable by its owner until the
// turn ends (lands included, which spend the land drop), and lost if it is not
// played. An empty library exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "623e261c-d4b9-4dea-930d-e0573a1deaee",
		Name:         "Grotag Night-Runner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Grotag Night-Runner — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					_, err := effects.B12ImpulseExileForTurn(g, item, 1)
					return err
				}),
		},
	})
}
