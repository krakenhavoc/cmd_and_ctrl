package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Storm Fleet Negotiator — Creature — Siren Pirate {2}{U}, 2/2
// (Lost Caverns of Ixalan Commander):
//
//	"Flying
//	 Parley — Whenever this creature attacks, each player reveals the top
//	 card of their library. For each nonland card revealed this way, you
//	 create a Map token. Then each player draws a card."
func init() {
	Register(Spec{
		OracleID:        "60cb885c-5cd3-4963-8a6a-cebf07cc61b5",
		Name:            "Storm Fleet Negotiator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Draws: 1, Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(stormFleetNegotiatorLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return cmdr2S49Parley(g, ctx.Source(), stormFleetNegotiatorLabel, func(nonland int) error {
					if nonland == 0 {
						return nil
					}
					return CreateToken{Template: MapToken(), N: nonland}.Apply(ctx)
				})
			}),
		},
	})
}

const stormFleetNegotiatorLabel = "Storm Fleet Negotiator — each player reveals the top card of their library; you create a Map for each nonland card; then each player draws a card"
