package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aeromoeba — Creature — Elemental Beast {3}{U}, 2/4:
//
//	"Flying
//	 Discard a card: Switch this creature's power and toughness until end
//	 of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1e276998-9c54-4d68-a901-657853312d06",
		Name:            "Aeromoeba",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Discard a card: Switch this creature's power and toughness until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.DiscardACard(),
				Effect:  effects.SwitchThisPT("Aeromoeba — switch its power and toughness"),
			},
		},
	})
}
