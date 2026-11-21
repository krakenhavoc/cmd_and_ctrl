package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Awaken the Sky Tyrant — Enchantment {3}{R} (Commander 2015, #5546):
//
//	"When a source an opponent controls deals damage to you, sacrifice this
//	 enchantment. If you do, create a 5/5 red Dragon creature token with
//	 flying."
//
// The trigger is Farsight Mask's condition (a source an opponent controls
// dealt damage to this permanent's controller), and the body is Promise of
// Bunrei's SacrificeThisThen: the Dragon comes only when the enchantment is
// actually sacrificed, so a second trigger that resolves after the first
// makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0607b289-74a6-4f3a-b15c-a639244873b6",
		Name:         "Awaken the Sky Tyrant",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ASourceAnOpponentControlsDealtDamageToYou,
				"Awaken the Sky Tyrant — sacrifice it; if you do, create a 5/5 red Dragon with flying",
				c15AwakenTheSkyTyrantSacrifice),
		},
	})
}

var c15AwakenTheSkyTyrantSacrifice = effects.SacrificeThisThen("then/awaken-the-sky-tyrant-dragon", func(ctx *effects.Context) error {
	return effects.CreateToken{Template: effects.TokenCard("5/5 red Dragon with flying"), N: 1}.Apply(ctx)
})
