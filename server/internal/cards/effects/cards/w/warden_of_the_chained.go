package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warden of the Chained — Creature — Minotaur Warrior {1}{R}{G}, 4/4 (Theros
// Beyond Death, #5514):
//
//	"Trample
//	 This creature can't attack unless you control another creature with power 4 or
//	 greater."
//
// An attack restriction read live at declaration over the controller's OTHER
// creatures (AttackTargetRestriction.ControllerMustControlAnother, #5514): the
// Warden's own 4 power never satisfies it (CR 109.1), and power is read after the
// layers (CR 613.4), so a pumped creature counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d2f33d3a-afc6-4d15-9049-ecd33f85eb94",
		Name:            "Warden of the Chained",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.Labelled(effects.CantAttackUnlessYouControlAnother(game.PermanentQuery{Types: []string{"creature"}}.WithPowerAtLeast(4)),
				"Can't attack unless you control another creature with power 4 or greater."),
		},
	})
}
