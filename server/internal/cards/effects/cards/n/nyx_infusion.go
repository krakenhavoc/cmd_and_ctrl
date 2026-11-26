package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nyx Infusion — Enchantment — Aura {2}{B} (Journey into Nyx):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 as long as it's an enchantment. Otherwise,
//	 it gets -2/-2."
//
// The host's type is read through the layered view, so an enchantment
// creature (a Theros god, or one a type-changing effect made so) gets the
// bonus and anything else the penalty, re-read on every recompute.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52e3bcc0-0c1f-4846-bf52-4af66b7abd6f",
		Name:         "Nyx Infusion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.B43PumpAttachedWhen(func(host *game.Card, g *game.Game, source *game.Card) bool {
				return effects.AttachedToSource(host, g, source) && host.IsEnchantment()
			}, 2, 2),
			effects.B43PumpAttachedWhen(func(host *game.Card, g *game.Game, source *game.Card) bool {
				return effects.AttachedToSource(host, g, source) && !host.IsEnchantment()
			}, -2, -2),
		},
	})
}
