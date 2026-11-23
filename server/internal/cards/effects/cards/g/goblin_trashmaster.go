package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Trashmaster — Creature — Goblin Warrior {2}{R}{R}, 3/3
// (EDHREC rank 3170):
//
//	"Other Goblins you control get +1/+1.
//	 Sacrifice a Goblin: Destroy target artifact."
//
// A Goblin lord with a sac outlet that eats artifacts. The anthem is
// the shared tribal builder over OTHER Goblins the controller
// controls; the sac outlet is Pashalik Mons' "Sacrifice a Goblin"
// cost — any Goblin the activator controls, the Trashmaster itself
// included — with the artifact destroyed on resolution if it is
// still there.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0283bf5e-ddf2-4a4a-a7cf-d3e27eed7e7d",
		Name:         "Goblin Trashmaster",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Goblin: Destroy target artifact",
			Cost:    effects.B25SacrificeAGoblin(),
			Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}
