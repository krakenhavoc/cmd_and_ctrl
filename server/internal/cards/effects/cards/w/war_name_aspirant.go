package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// War-Name Aspirant — Creature — Human Warrior {1}{R}, 2/1 (Jumpstart 2022,
// #5948):
//
//	"Raid — This creature enters with a +1/+1 counter on it if you attacked
//	 this turn.
//	 This creature can't be blocked by creatures with power 1 or less."
//
// Swaggering Corsair's raid replacement (CR 614.1c) plus Arlinn's Wolf's
// block restriction, with the power read through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "448c84f9-712b-4970-8719-2ab9c2a59d20",
		Name:         "War-Name Aspirant",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.AvB16SelfEntersWithCountersIfRaid(game.CounterPlusOne, 1)},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.PowerLE(1), "creatures with power 1 or less"),
		},
	})
}
