package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bat Whisperer — Creature — Vampire {3}{B}, 4/2 (Innistrad: Midnight Hunt,
// #5497):
//
//	"When this creature enters, if an opponent lost life this turn, create
//	 a 1/1 black Bat creature token with flying."
//
// Bloodtithe Collector's intervening "if" (CR 603.4), checked as it would
// trigger and again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f20f6da9-d678-486b-99f9-7121669fef81",
		Name:         "Bat Whisperer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Bat Whisperer — create a 1/1 Bat with flying (an opponent lost life this turn)",
				effects.Do(effects.If{
					Cond: effects.AvB34CondAnOpponentLostLife{},
					Then: []effects.Applier{effects.CreateToken{Template: effects.TokenCard("1/1 black Bat with flying"), N: 1}},
				})),
		},
	})
}
