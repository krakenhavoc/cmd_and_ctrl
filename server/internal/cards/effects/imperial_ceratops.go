package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperial Ceratops — Creature — Dinosaur {4}{W}, 3/5 (Rivals of Ixalan,
// #5531):
//
//	"Enrage — Whenever this creature is dealt damage, you gain 2 life."
//
// Dromad Purebred's enrage at 2: one trigger per damage instance, however
// many sources dealt it at once (CR 120.3), and lethal damage still
// triggers it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "86cee7f0-0aac-46f1-8e4e-062fc398477e",
		Name:         "Imperial Ceratops",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Imperial Ceratops — Enrage: you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
