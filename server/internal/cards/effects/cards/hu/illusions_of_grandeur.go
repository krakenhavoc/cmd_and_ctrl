package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Illusions of Grandeur — Enchantment {3}{U} (Ice Age, S82 OLD batch 839, #5636):
//
//	"Cumulative upkeep {2}
//	 When this enchantment enters, you gain 20 life.
//	 When this enchantment leaves the battlefield, you lose 20 life."
//
// Two ordinary triggers. The leaves trigger is controlled by the player who
// controlled the enchantment as it left (CR 603.10a, last-known information), and
// they lose the life even though it is gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7c81b3e-8082-4aaf-8979-2717d118234c",
		Name:         "Illusions of Grandeur",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Illusions of Grandeur — cumulative upkeep {2}", "{2}"),
			effects.WhenThisEnters("Illusions of Grandeur — you gain 20 life", effects.Do(effects.GainLife{Amount: 20})),
			effects.WhenThisLeaves("Illusions of Grandeur — you lose 20 life", effects.Do(effects.YouLoseLife(20))),
		},
	})
}
