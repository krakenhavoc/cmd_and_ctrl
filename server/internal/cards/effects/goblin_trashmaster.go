package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0283bf5e-ddf2-4a4a-a7cf-d3e27eed7e7d",
		Name:         "Goblin Trashmaster",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Goblin"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Goblin: Destroy target artifact",
			Cost:    b25SacrificeAGoblin(),
			Targets: TargetPermanent("target artifact", Artifact()),
			Effect:  b17DestroyFirstLegalTarget,
		}},
	})
}
