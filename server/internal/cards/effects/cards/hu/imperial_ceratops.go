package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "86cee7f0-0aac-46f1-8e4e-062fc398477e",
		Name:         "Imperial Ceratops",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Cmdr2B08EnrageTrigger("Imperial Ceratops — Enrage: you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
