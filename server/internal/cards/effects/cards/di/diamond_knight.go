package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diamond Knight — Artifact Creature — Knight {3}, 1/1 (Core Set 2020, #5518):
//
//	"Vigilance
//	 As this creature enters, choose a color.
//	 Whenever you cast a spell of the chosen color, put a +1/+1 counter on this
//	 creature."
//
// Prism Ring's shape: the colour is chosen as it enters (CR 614.12), and the
// trigger reads the colours of the spell its controller casts; the counter goes
// on this creature if it is still the same permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "29ae4471-c28f-4e43-8afb-fee5023e4d08",
		Name:            "Diamond Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		AsEnters:        effects.ChooseColorAsEnters(game.ColorForBenefit, "Diamond Knight"),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastASpellOfTheChosenColor,
				"Diamond Knight — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
