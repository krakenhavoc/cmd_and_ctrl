package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Broken Sky — Enchantment, the BACK FACE of Invasion of Tolvada
// (face 1):
//
//	"Creature tokens you control get +1/+0 and have lifelink.
//	 At the beginning of your end step, create a 1/1 white and black
//	 Spirit creature token with flying."
//
// Registered under "<oracle_id>#1". Both statics reach only creature
// tokens, the Spirit it makes included.
//
// No simplification.
func init() {
	lifelink := effects.KeywordGrant(effects.B05CreatureTokenYouControl, "lifelink")
	lifelink.Label = "Creature tokens you control have lifelink."
	effects.Register(effects.Spec{
		OracleID:     effects.InvasionOfTolvadaOracleID + "#1",
		Name:         "The Broken Sky",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "Creature tokens you control get +1/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.B05CreatureTokenYouControl,
				Apply:     effects.PumpBy(1, 0),
			},
			lifelink,
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("The Broken Sky — create a 1/1 white and black Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white and black Spirit with flying"), N: 1})),
		},
	})
}
