package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Turtleshell Changeling — Creature — Shapeshifter {3}{U}, 1/4:
//
//	"Changeling (This card is every creature type.)
//	 {1}{U}: Switch this creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "87f8b403-765c-4a2a-8b21-edaaaa889234",
		Name:            "Turtleshell Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"changeling"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{U}: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{1}{U}"),
				Effect:  effects.SwitchThisPT("Turtleshell Changeling — switch its power and toughness"),
			},
		},
	})
}
