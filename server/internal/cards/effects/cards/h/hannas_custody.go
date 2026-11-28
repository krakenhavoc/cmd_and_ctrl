package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hanna's Custody — Enchantment {2}{W} (Tempest, S82 OLD batch 765, #5628):
//
//	"All artifacts have shroud. (They can't be the targets of spells or
//	 abilities.)"
//
// A layer 6 keyword grant over every artifact on the battlefield, whoever
// controls it (CR 613.1f). Shroud is read through the layered view, so an
// artifact that has lost its abilities still has the shroud this gives it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57b0205d-ad9d-45b7-8556-0733aa7a4987",
		Name:         "Hanna's Custody",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsArtifact()
			}, "shroud"),
		},
	})
}
