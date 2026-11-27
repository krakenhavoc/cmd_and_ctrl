package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wu Admiral — Creature — Human Soldier {4}{U}, 3/3
// (Portal Three Kingdoms, #5622):
//
//	"This creature gets +1/+1 as long as an opponent controls an Island."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "578c5b92-9f7e-4645-9624-f2c1a16b6aaf",
		Name:         "Wu Admiral",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as an opponent controls an Island.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.OldB710SelfWhileAnOpponentControlsA(effects.Subtype("Island")),
			Apply:     effects.PumpBy(1, 1),
		}},
	})
}
