package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subterranean Schooner — {1}{U} Artifact — Vehicle, 3/4:
//
//	"Whenever this Vehicle attacks, target creature that crewed it this
//	 turn explores.
//	 Crew 1"
//
// The target clause is CrewedThis (ADR 0186 section 3). The explorer is
// read from the announced target at resolution (CR 608.2b), and the
// explore itself is the engine's CR 701.44 action under the trigger's
// controller.
//
// No simplification.
func init() {
	attack := effects.WheneverThisAttacks("Subterranean Schooner — target creature that crewed it explores", subterraneanSchoonerExplore)
	attack.Targets = effects.CrewedThis(effects.TargetCreature("target creature that crewed it this turn"))

	effects.Register(effects.Spec{
		OracleID:     "7f82bae3-feaa-4967-a843-32a4f72806d5",
		Name:         "Subterranean Schooner",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Subterranean Schooner"),
		}},
		Triggered: []game.TriggeredAbility{attack},
	})
}

func subterraneanSchoonerExplore(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.Explores{Explorer: id}.Apply(ctx)
}
