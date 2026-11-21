package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mummy Paramount — Creature — Zombie {1}{W}, 2/2 (Hour of Devastation,
// #5534):
//
//	"Whenever another Zombie you control enters, this creature gets +1/+1
//	 until end of turn."
//
// Wayward Servant's trigger condition (another permanent with the Zombie
// subtype entering under your control, a token included) on the shared
// thisGetsUntilEndOfTurn body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99cc909d-3cb2-499c-8231-41418745cd71",
		Name:         "Mummy Paramount",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Zombie"))
			}, "Mummy Paramount — +1/+1 until end of turn", effects.ThisGetsUntilEndOfTurn(1, 1, "Mummy Paramount — +1/+1 until end of turn")),
		},
	})
}
