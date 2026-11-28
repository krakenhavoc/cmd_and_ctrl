package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fountain Watch — Creature — Human Cleric {3}{W}{W}, 2/4 (Mercadian Masques, S82 OLD batch 688, #5619):
//
//	"Artifacts and enchantments you control have shroud.
//	(They can't be the targets of spells or abilities.)"
//
// A layer-6 keyword grant (CR 613.1f) of shroud to every artifact and enchantment its controller
// controls (KeywordGrant), the Watch aside, since it is neither; an artifact creature or
// enchantment creature counts. Shroud is the engine's own keyword, so targeting reads it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6527ba3-3293-4a29-a2f9-7a81f8f27b7c",
		Name:         "Fountain Watch",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.Controller == source.Controller && (target.IsArtifact() || target.IsEnchantment())
			}, "shroud"),
		},
	})
}
