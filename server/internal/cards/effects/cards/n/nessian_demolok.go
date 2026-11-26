package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nessian Demolok — {3}{G}{G} Creature — Beast 3/3 (Born of the Gods):
//
//	"Tribute 3 (As this creature enters, an opponent of your choice may put three
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, destroy target noncreature
//	 permanent."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2341b59b-ce46-4efe-be94-c3851d3cd9e8",
		Name:         "Nessian Demolok",
		Completeness: effects.CompletenessFull,
		Replacements: effects.Tribute("Nessian Demolok", 3),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenEntersIfTributeWasntPaid("Nessian Demolok — destroy target noncreature permanent",
				effects.DestroyFirstLegalTarget),
				effects.TargetPermanent("target noncreature permanent", effects.Noncreature())),
		},
	})
}
