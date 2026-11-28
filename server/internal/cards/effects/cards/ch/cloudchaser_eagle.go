package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cloudchaser Eagle — Creature — Bird {3}{W}, 2/2 (Tempest, S82 OLD, #5628):
//
//	"Flying
//	 When this creature enters, destroy target enchantment."
//
// A mandatory enters trigger, as Indrik Stomphowler's: with no enchantment
// anywhere it has no target and is removed (CR 603.3d); it may target the
// controller's own. Flying is read from the printed card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "add18787-2329-4355-a524-66fefda8d06e",
		Name:            "Cloudchaser Eagle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Cloudchaser Eagle — destroy target enchantment", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target enchantment", effects.Enchantment())),
		},
	})
}
