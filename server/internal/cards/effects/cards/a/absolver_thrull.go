package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Absolver Thrull — Creature — Thrull Cleric {3}{W}, 2/3 (Guildpact):
//
//	"Haunt (When this creature dies, exile it haunting target creature.)
//	 When this creature enters or the creature it haunts dies, destroy
//	 target enchantment."
//
// Haunt is CR 702.55 (#6083): one ability with two conditions, the second
// met in exile (CR 702.55c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9002245-0f33-48cb-96d3-fde884236eb5",
		Name:         "Absolver Thrull",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Haunt("Absolver Thrull"),
			effects.Targeting(
				effects.WhenThisEntersOrTheCreatureItHauntsDies("Absolver Thrull — destroy target enchantment", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
			),
		},
	})
}
