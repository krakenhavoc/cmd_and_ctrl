package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Floral Spuzzem — Creature — Elemental {3}{G}, 2/2:
//
//	"Whenever this creature attacks and isn't blocked, you may destroy target artifact defending
//	 player controls. If you do, this creature assigns no combat damage this turn."
//
// #3531: the artifact is targeted when the trigger goes on the stack; choosing to destroy it is
// the "if you do".
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "994de451-14f9-466f-a56e-da052b4666e5",
		Name:         "Floral Spuzzem",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(unblockedAttackerMay("Floral Spuzzem", "destroy target artifact defending player controls", "floral-spuzzem-destroy",
				func(ctx *Context, target uuid.UUID) error {
					return DestroyTarget{Target: target}.Apply(ctx)
				}), TargetOfDefendingPlayer("artifact", "target artifact defending player controls", Artifact())),
		},
	})
}
