package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seasoned Warrenguard — Creature — Rabbit Warrior {W}, 1/2 (Bloomburrow,
// #4467):
//
//	"Whenever this creature attacks while you control a token, this
//	 creature gets +2/+0 until end of turn."
//
// Brazen Blademaster's shape: "while you control a token" is part of the
// trigger event, not an intervening "if" (CR 603.4 needs the word "if"),
// so it is read once, as the attack is declared, over any token
// permanent you control (controlsAnyCard over IsTokenPredicate). A token
// lost in response does not stop the +2/+0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bf1dbb6-ed65-4968-8edd-b98e9bd19d4d",
		Name:         "Seasoned Warrenguard",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclared(ev, source) && effects.ControlsAnyCard(g, source.Controller, effects.IsTokenPredicate())
			}, "Seasoned Warrenguard — gets +2/+0 until end of turn", effects.ThisGetsUntilEndOfTurn(2, 0, "Seasoned Warrenguard")),
		},
	})
}
