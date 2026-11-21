package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torment of Scarabs — Enchantment — Aura Curse {3}{B} (Foundations Jumpstart):
//
//	"Enchant player
//	 At the beginning of enchanted player's upkeep, that player loses 3 life
//	 unless they sacrifice a nonland permanent of their choice or discard a card."
//
// The enchanted player picks: lose 3 life, sacrifice a nonland permanent,
// or discard a card. Options they can't take are not offered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ba5bcee-fc95-4d05-965e-b1762529ec5f",
		Name:         "Torment of Scarabs",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.AttachedTo.Kind == game.TargetPlayer && source.AttachedTo.ID == ev.Actor
			}, "Torment of Scarabs — that player loses 3 life unless they sacrifice a nonland permanent or discard a card",
				effects.TormentOfScarabsUpkeep),
		},
	})
}
