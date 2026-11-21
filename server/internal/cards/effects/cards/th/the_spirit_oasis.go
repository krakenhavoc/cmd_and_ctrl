package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Spirit Oasis — Legendary Enchantment — Shrine {2}{U} (Avatar: The
// Last Airbender, #4459):
//
//	"When The Spirit Oasis enters, draw a card for each Shrine you
//	 control.
//	 Whenever another Shrine you control enters, draw a card."
//
// The Shrines are counted as the enters trigger resolves, The Spirit
// Oasis among them while it is still on the battlefield (Sanctum of
// Fruitful Harvest's count). The second trigger watches every other
// Shrine entering under your control, tokens and copies included, one
// card each.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b677a73c-f179-46bf-83ee-c2b05f9fca67",
		Name:         "The Spirit Oasis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("The Spirit Oasis — draw a card for each Shrine you control", func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{N: effects.AvatarB6ShrinesYouControl(g, item.Controller)}.Apply(effects.NewContext(g, item))
			}),
			effects.On(game.EventETB, effects.AvatarB6AnotherShrineEnteredUnderYourControl,
				"The Spirit Oasis — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
