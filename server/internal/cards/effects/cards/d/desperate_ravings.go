package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Ravings — Instant {1}{R} (Commander 2019):
//
//	"Draw two cards, then discard a card at random.
//	 Flashback {2}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Faithless Looting's flashback with Aether Rift's random discard: the two
// draws land first, so a drawn card can be the one discarded, and the discard
// is chosen by the engine's random draw, not the player. Flashback is the
// shared constructor, whose cost is a different colour from the card's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "8b96c334-82c7-4511-805e-f04c27514d7c",
		Name:             "Desperate Ravings",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Draws: 2, Discards: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{U}")},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.DrawCards{N: 2}).Apply(ctx); err != nil {
				return err
			}
			return ctx.Game.DiscardRandomKeyedForEffect(item.Controller, 1, game.KeyedTail{})
		},
	})
}
