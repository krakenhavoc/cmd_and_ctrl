package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rite of Harmony — Instant {G}{W} (Innistrad: Midnight Hunt, #5497):
//
//	"Whenever a creature or enchantment you control enters this turn,
//	 draw a card.
//	 Flashback {2}{G}{W}"
//
// The first sentence creates a delayed triggered ability (CR 603.7a) that
// lasts the turn (CR 603.7b) and triggers for each creature or enchantment
// that enters under your control (CR 603.2c), the First Day of Class shape
// with a wider condition. A permanent that entered before Rite of Harmony
// resolved does not trigger it, and a permanent that is both (an enchantment
// creature) triggers it once. Two copies resolved (the flashback) draw twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "af94fd87-ea37-49eb-b5e6-0b1846f23859",
		Name:             "Rite of Harmony",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{2}{G}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DelayedOnEvent{
				Label:     "Rite of Harmony — draw a card",
				On:        []game.EventKind{game.EventETB},
				Condition: effects.S82Mid10CreatureOrEnchantmentEntersCondition,
				Body:      effects.S82Mid10DrawACardBody,
				Repeats:   true,
			}.Apply(ctx)
		},
	})
}
