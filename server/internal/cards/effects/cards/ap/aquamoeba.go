package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aquamoeba — Creature — Elemental Beast {1}{U}, 1/3:
//
//	"Discard a card: Switch this creature's power and toughness until end
//	 of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55de434c-0fb9-4127-81b6-47c1c0b088fe",
		Name:         "Aquamoeba",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Discard a card: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.DiscardACard(),
				Effect:  effects.SwitchThisPT("Aquamoeba — switch its power and toughness"),
			},
		},
	})
}
