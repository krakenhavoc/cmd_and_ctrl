package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dakmor Lancer — Creature — Human Knight {4}{B}{B}, 3/3 (Starter 1999,
// #5620):
//
//	"When this creature enters, destroy target nonblack creature."
//
// Dark Hatchling's enters trigger without "It can't be regenerated": the
// target is judged as the trigger goes on the stack and again as it
// resolves (CR 608.2b), so a creature that turned black in response is
// spared, and a regeneration shield on the target saves it. With no
// nonblack creature on the battlefield the trigger is removed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00efb113-ba5a-4749-b42b-5693b980f26f",
		Name:         "Dakmor Lancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Dakmor Lancer — destroy target nonblack creature", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target nonblack creature", effects.NonBlack())),
		},
	})
}
