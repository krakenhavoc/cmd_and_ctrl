package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arrogant Outlaw — Creature — Vampire Noble {2}{B}, 3/2 (Innistrad: Midnight
// Hunt, #5497):
//
//	"When this creature enters, if an opponent lost life this turn, each
//	 opponent loses 2 life and you gain 2 life."
//
// Bloodtithe Collector's intervening "if" (CR 603.4) with Brotherhood
// Patriarch's drain.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f202712d-7cd1-47aa-9664-526a884a1be2",
		Name:         "Arrogant Outlaw",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Arrogant Outlaw — each opponent loses 2 life and you gain 2 life (an opponent lost life this turn)",
				effects.Do(effects.If{
					Cond: effects.S82MidB06CondAnOpponentLostLife{},
					Then: []effects.Applier{effects.EachOpponentLosesLife{N: 2}, effects.GainLife{Amount: 2}},
				})),
		},
	})
}
