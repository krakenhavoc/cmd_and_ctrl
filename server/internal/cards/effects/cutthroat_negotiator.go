package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cutthroat Negotiator — Creature — Orc Pirate {2}{U}{R}, 4/3 (March of
// the Machine Commander, #5477):
//
//	"Parley — Whenever this creature attacks, each player reveals the top
//	 card of their library. For each nonland card revealed this way, you
//	 create a tapped Treasure token. Then each player draws a card."
//
// Storm Fleet Negotiator's parley (cmdr2S49Parley) with Goldvein Hydra's
// tapped Treasures (b13CreateTappedTreasures) as the reward: every
// player reveals in turn order, the Treasures come before the draws, and
// all of them are the attacker's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af14101a-9cc0-4bd3-a9f9-2726353e40ff",
		Name:         "Cutthroat Negotiator",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(cutthroatNegotiatorLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return cmdr2S49Parley(g, ctx.Source(), cutthroatNegotiatorLabel, func(nonland int) error {
					return b13CreateTappedTreasures(ctx, ctx.Controller(), nonland)
				})
			}),
		},
	})
}

const cutthroatNegotiatorLabel = "Cutthroat Negotiator — each player reveals the top card of their library; you create a tapped Treasure for each nonland card; then each player draws a card"
