package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reliquary Monk — Creature — Human Monk Cleric {2}{W}, 2/2 (Urza's
// Destiny, #5621):
//
//	"When this creature dies, destroy target artifact or enchantment."
//
// Mudbutton Cursetosser's targeted dies trigger with a Disenchant
// target: any player's artifact or enchantment, the Monk's controller's
// own included. The target is chosen as the trigger goes on the stack
// and checked again as it resolves; if it has left, nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a149ac31-b792-4fbe-903a-788bbdfd5e97",
		Name:         "Reliquary Monk",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Reliquary Monk — destroy target artifact or enchantment", effects.DestroyFirstLegalTarget),
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			),
		},
	})
}
