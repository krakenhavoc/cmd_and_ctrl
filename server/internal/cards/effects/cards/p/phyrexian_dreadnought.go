package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Dreadnought — Artifact Creature — Phyrexian Dreadnought {1},
// 12/12 (Mirage, #5633):
//
//	"Trample
//	 When this creature enters, sacrifice it unless you sacrifice any number
//	 of creatures with total power 12 or greater."
//
// The pay-unless prompt with the total-power sacrifice payment (#6187): the
// controller names any number of creatures they control (the Dreadnought
// itself among them) with total power 12 or more, or declines, and the
// Dreadnought is sacrificed. With less than 12 power on their side it is
// sacrificed without a question.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7950edb-5670-46fd-a30e-0fe888a5a32b",
		Name:            "Phyrexian Dreadnought",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.EntersSacrificeUnlessPaying("Phyrexian Dreadnought",
				effects.SacrificeTotalPowerPayment(12, "creatures", effects.QueryType("creature"))),
		},
	})
}
