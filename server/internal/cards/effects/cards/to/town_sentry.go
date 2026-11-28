package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Town Sentry — Creature — Human Soldier {2}{W}, 2/2 (Portal Second Age, S82 older sets batch 735, #5625):
//
//	"Whenever this creature blocks, it gets +0/+2 until end of turn."
//
// Triggers once when it blocks, however many attackers it blocks (CR 509.3a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec1ac4df-b031-43b4-a2de-ca58a1599053",
		Name:         "Town Sentry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.SelfBlocksOnce(ev, source)
			}, "Town Sentry — it gets +0/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(0, 2, "Town Sentry — +0/+2")),
		},
	})
}
