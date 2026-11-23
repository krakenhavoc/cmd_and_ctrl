package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marshal of Zhalfir — Creature — Human Knight {W}{U}, 2/2 (March of the
// Machine):
//
//	"Other Knights you control get +1/+1.
//	 {W}{U}, {T}: Tap another target creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ad01fb2-2ba7-447a-a5ef-e57d46421635",
		Name:         "Marshal of Zhalfir",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Knight"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{U}, {T}: Tap another target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}{U}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature")),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
