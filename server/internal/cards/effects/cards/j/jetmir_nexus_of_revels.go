package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jetmir, Nexus of Revels — Legendary Creature — Cat Demon
// {1}{R}{G}{W}, 5/4 (EDHREC rank 2221):
//
//	"Creatures you control get +1/+0 and have vigilance as long as
//	 you control three or more creatures.
//	 Creatures you control also get +1/+0 and have trample as long as
//	 you control six or more creatures.
//	 Creatures you control also get +1/+0 and have double strike as
//	 long as you control nine or more creatures."
//
// The go-wide payoff, three thresholds. Each line is a pair of
// statics (b20JetmirClause): a layer 7c +1/+0 and a layer 6 keyword
// grant, both gated on the live creature count of the controller —
// Jetmir included, as printed — so the bonuses switch on and off as
// creatures enter and leave. At nine the three lines stack to +3/+0
// with vigilance, trample and double strike.
//
// No simplification.
func init() {
	var statics []game.StaticAbility
	statics = append(statics, effects.B20JetmirClause(3, "vigilance")...)
	statics = append(statics, effects.B20JetmirClause(6, "trample")...)
	statics = append(statics, effects.B20JetmirClause(9, "double strike")...)
	effects.Register(effects.Spec{
		OracleID:     "da72a4bc-ce6f-4b72-bc66-2ee33cfa87df",
		Name:         "Jetmir, Nexus of Revels",
		Completeness: effects.CompletenessFull,
		Static:       statics,
	})
}
