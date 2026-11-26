package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spectra Ward — Enchantment — Aura {3}{W}{W} (Magic 2015, #5556):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has protection from each color. This
//	 effect doesn't remove Auras."
//
// A layer 7c pump and a layer 6 grant of the five protections, each minted as
// "doesn't remove Auras" (CR 702.16n): this Aura, and every other Aura on the
// creature, stays attached, though they are coloured. An Aura put on it later
// stays too; Equipment does not, as the card says only Auras.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "effb99e6-db1b-4ada-879d-431d9d13d4b7",
		Name:         "Spectra Ward",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			{
				Label:     "Enchanted creature has protection from each color. This effect doesn't remove Auras.",
				Layer:     game.Layer6Ability,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, col := range game.AllColors {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, game.ProtectionSparingAuras(game.ProtectionFromColor(col)))
					}
				},
			},
		},
	})
}
