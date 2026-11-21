package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cacophodon — Creature — Dinosaur {3}{G}, 2/5 (Rivals of Ixalan, #5531):
//
//	"Enrage — Whenever this creature is dealt damage, untap target
//	 permanent."
//
// The shared enrage trigger with Voltaic Key's untap body: any permanent,
// picked as the trigger goes on the stack (CR 603.3d) and re-checked as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bd6dea2-2fdd-482f-9fe8-6f3eeed13872",
		Name:         "Cacophodon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Cmdr2B08EnrageTrigger("Cacophodon — Enrage: untap target permanent", effects.UntapFirstLegalTarget),
				effects.TargetPermanent("target permanent")),
		},
	})
}
