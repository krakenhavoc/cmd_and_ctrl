package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dance of the Skywise — Instant {1}{U} (Dragons of Tarkir):
//
//	"Until end of turn, target creature you control becomes a blue Dragon Illusion
//	 with base power and toughness 4/4, loses all abilities, and gains flying."
//
// One record: its colour is set to blue, its creature types to Dragon Illusion
// (CR 205.1a), its base size to 4/4, it loses every ability and then has flying
// (CR 613.1f; the keep list of the loss is the grant that follows it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ab0ab9b-c143-4315-8983-f645fb21f8b3",
		Name:         "Dance of the Skywise",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve:    effects.ResolveSteps(effects.OnTheTargetCard(effects.DanceOfTheSkywiseBecomes)),
	})
}
