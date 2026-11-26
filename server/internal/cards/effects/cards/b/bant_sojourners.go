package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bant Sojourners — Creature — Human Soldier {1}{G}{W}{U}, 2/4 (Alara Reborn,
// #5586):
//
//	"When you cycle this card and when this creature dies, you may create a
//	 1/1 white Soldier creature token.
//	 Cycling {2}{W}"
//
// One ability with two conditions (S82OldB501CycledOrDies); the token is
// optional and is created under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67ceffa4-2fdb-499c-88cd-49fb5eb9be59",
		Name:         "Bant Sojourners",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.S82OldB501CycledOrDies("Bant Sojourners — you may create a 1/1 white Soldier creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: 1})),
				"Bant Sojourners — create a 1/1 white Soldier token?"),
		},
	})
}
