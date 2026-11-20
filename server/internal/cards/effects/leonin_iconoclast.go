package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "14d14ded-8831-47ad-ad90-f6f242cccf03",
		Name:         "Leonin Iconoclast",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(heroic("Leonin Iconoclast — destroy target enchantment creature an opponent controls (heroic)", destroyFirstLegalTarget),
				TargetCreature("target enchantment creature an opponent controls", Enchantment(), OpponentControls())),
		},
	})
}
