package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Felidar Guardian — Creature — Cat Beast {3}{W}, 1/4:
//
//	"When this creature enters, you may exile another target permanent you
//	 control, then return that card to the battlefield under its owner's
//	 control."
//
// Restoration Angel's blink with the clause widened to any other permanent you
// control and the return under its owner's control. "Another" is the object,
// never a name. A token exiled this way ceases to exist; a permanent that has
// left by resolution is skipped (CR 608.2b). The blinked card is a new object
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33869ba6-13e5-4e48-8963-92f7965648fb",
		Name:         "Felidar Guardian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.WhenThisEnters("Felidar Guardian — exile another target permanent you control, then return that card to the battlefield under its owner's control",
						effects.FlickerFirstLegalTargetUnderOwner),
					effects.Another(effects.TargetPermanent("another target permanent you control", effects.YouControl()))),
				"Felidar Guardian — exile another target permanent you control, then return it to the battlefield?"),
		},
	})
}
