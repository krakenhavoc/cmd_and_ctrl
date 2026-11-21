package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clarion Spirit — Creature — Spirit {1}{W}, 2/2 (Kaldheim, #5505):
//
//	"Whenever you cast your second spell each turn, create a 1/1 white
//	 Spirit creature token with flying."
//
// Wingblade Disciple's trigger (YouCastYourSecondSpellEachTurn) with a
// Spirit token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dec20b18-afe6-405e-b47d-a757a5645c63",
		Name:         "Clarion Spirit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Clarion Spirit — create a 1/1 white Spirit creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 1})),
		},
	})
}
