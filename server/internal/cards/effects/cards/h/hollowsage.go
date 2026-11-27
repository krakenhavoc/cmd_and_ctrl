package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hollowsage — Creature — Merfolk Wizard {3}{B}, 2/2 (Shadowmoor, #5591):
//
//	"Whenever this creature becomes untapped, you may have target player
//	 discard a card."
//
// One trigger each time the Hollowsage itself untaps, in the untap step or by
// an effect. The target player is chosen as the trigger goes on the stack
// (CR 603.3d) and the "you may" is asked as it resolves (CR 603.5); the
// player chosen picks the card to discard, and an empty hand discards
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e1efd99-8bfc-4650-a430-a754578facbd",
		Name:         "Hollowsage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.On(game.EventUntapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventUntapCard && ev.CardID == source.InstanceID
			}, "Hollowsage — you may have target player discard a card", effects.TargetOpponentDiscards(1)),
				"Hollowsage — have the target player discard a card?"), effects.TargetPlayer("target player")),
		},
	})
}
