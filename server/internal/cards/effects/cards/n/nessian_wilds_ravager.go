package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nessian Wilds Ravager — {4}{G}{G} Creature — Hydra 6/6 (Born of the Gods):
//
//	"Tribute 6 (As this creature enters, an opponent of your choice may put six
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, you may have this creature
//	 fight another target creature."
//
// The target is chosen as the trigger goes on the stack; "you may" is asked as it
// resolves (Affectionate Indrik's shape).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa1a5797-3076-48a4-b64b-ce3dfe2a3e7b",
		Name:         "Nessian Wilds Ravager",
		Completeness: effects.CompletenessFull,
		Replacements: effects.Tribute("Nessian Wilds Ravager", 6),
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenEntersIfTributeWasntPaid("Nessian Wilds Ravager — it fights another target creature",
				effects.B30SourceFightsFirstLegalTarget),
				effects.Another(effects.TargetCreature("another target creature"))),
				"Nessian Wilds Ravager — have it fight another target creature?"),
		},
	})
}
