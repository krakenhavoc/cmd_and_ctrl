package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Voltaic Servant — Artifact Creature — Construct {2}, 1/3 (Dominaria, S82 OLD
// batch 215, #5530):
//
//	"At the beginning of your end step, untap target artifact."
//
// A trigger at the beginning of your own end step with a target artifact
// chosen as it goes on the stack, re-checked as it resolves (CR 608.2b). The
// Servant is an artifact itself, so it can untap itself; the artifact may be
// anyone's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0c2703b-bd74-41ac-b516-c46c654de1bb",
		Name:         "Voltaic Servant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.AtYourEndStep("Voltaic Servant — untap target artifact", effects.UntapEachLegalTarget),
				effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}
