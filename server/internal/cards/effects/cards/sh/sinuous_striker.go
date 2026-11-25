package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sinuous Striker — {2}{U} Creature — Snake Warrior 2/2 (Hour of Devastation):
//
//	"{U}: This creature gets +1/-1 until end of turn.
//	 Eternalize—{3}{U}{U}, Discard a card."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4ed24b2-ba4f-492d-8b8d-44a5b21e4b33",
		Name:         "Sinuous Striker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.ThisGetsRow("{U}", 1, -1, "Sinuous Striker"),
			effects.OldB238EternalizeDiscard("{3}{U}{U}"),
		},
	})
}
