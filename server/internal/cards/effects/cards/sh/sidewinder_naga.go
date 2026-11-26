package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sidewinder Naga — {2}{G} Creature — Snake Warrior 3/2 (Hour of Devastation):
//
//	"As long as you control a Desert or there is a Desert card in your
//	 graveyard, this creature gets +1/+0 and has trample."
//
// Two statics under one condition: the pump in layer 7c and the keyword in
// layer 6.
func init() {
	applies := effects.OldB240SelfWhile(effects.OldB240YouHaveADesert)
	effects.Register(effects.Spec{
		OracleID:     "92247856-35a4-4c1d-90bc-fa0d28524ecd",
		Name:         "Sidewinder Naga",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "As long as you control a Desert or there is a Desert card in your graveyard, this creature gets +1/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: applies,
				Apply:     effects.PumpBy(1, 0),
			},
			effects.KeywordGrant(applies, "trample"),
		},
	})
}
