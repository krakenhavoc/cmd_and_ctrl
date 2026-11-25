package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sage of Mysteries — Creature — Human Wizard {U}, 0/2:
//
//	"Constellation — Whenever an enchantment you control enters, target player
//	 mills two cards."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62c949c8-0011-4ab7-8388-e643be79df83",
		Name:         "Sage of Mysteries",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.GapB103EnchantmentEnteredUnderYourControl,
					"Sage of Mysteries — target player mills two cards (constellation)",
					func(g *game.Game, item *game.StackItem) error {
						for _, t := range effects.NewContext(g, item).LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return effects.MillCards{Player: t.ID, N: 2}.Apply(effects.NewContext(g, item))
							}
						}
						return nil
					}),
				effects.TargetPlayer("target player")),
		},
	})
}
