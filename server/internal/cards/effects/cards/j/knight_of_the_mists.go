package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of the Mists — Creature — Human Knight {2}{U}, 2/2 (Visions, S82
// older sets batch 790, #5631):
//
//	"Flanking
//	 When this creature enters, you may pay {U}. If you don't, destroy target
//	 Knight and it can't be regenerated."
//
// The target Knight is chosen as the trigger goes on the stack, and may be the
// Knight of the Mists itself. At resolution a target that is no longer legal
// cancels the whole ability (CR 608.2b); otherwise the controller is asked to
// pay {U}, and declining (or being unable to pay) destroys the Knight with
// regeneration off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9893534-3d05-42c2-93ff-f28ffae89aa3",
		Name:            "Knight of the Mists",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flanking"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Knight of the Mists — you may pay {U}; if you don't, destroy target Knight",
				effects.Old790KnightOfTheMistsETB), effects.TargetCreature("target Knight", effects.HasSubtype("Knight"))),
		},
	})
}
