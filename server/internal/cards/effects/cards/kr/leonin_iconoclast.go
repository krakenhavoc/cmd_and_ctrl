package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leonin Iconoclast — Creature — Cat Monk {3}{W}, 3/2 (Journey into Nyx, #5558):
//
//	"Heroic — Whenever you cast a spell that targets this creature,
//	 destroy target enchantment creature an opponent controls."
//
// Arena Athlete's targeted heroic, destroying. With no enchantment
// creature an opponent controls, the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14d14ded-8831-47ad-ad90-f6f242cccf03",
		Name:         "Leonin Iconoclast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Heroic("Leonin Iconoclast — destroy target enchantment creature an opponent controls (heroic)", effects.DestroyFirstLegalTarget),
				effects.TargetCreature("target enchantment creature an opponent controls", effects.Enchantment(), effects.OpponentControls())),
		},
	})
}
