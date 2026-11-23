package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wispweaver Angel — Creature — Angel {4}{W}{W}, 4/4 (Jumpstart 2022):
//
//	"Flying
//	 When this creature enters, you may exile another target creature you
//	 control, then return that card to the battlefield under its owner's
//	 control."
//
// Restoration Angel's flicker without the Angel restriction and with "another"
// (object identity, so the Angel itself is never a legal pick). The target is
// chosen as the trigger goes on the stack and the "you may" is asked as it
// resolves; the creature returns as a new object under its owner's control
// (CR 400.7, 110.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8a88521-80e5-419b-ab4e-f4b3aba8afa0",
		Name:            "Wispweaver Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.WhenThisEnters("Wispweaver Angel — exile another target creature you control, then return it to the battlefield",
						effects.FlickerFirstLegalTarget),
					effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				"Wispweaver Angel — exile another creature you control, then return it to the battlefield?"),
		},
	})
}
