package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// River Sneak — Creature — Merfolk Warrior {1}{U}, 1/1 (Jumpstart 2022, #5948):
//
//	"This creature can't be blocked.
//	 Whenever another Merfolk you control enters, this creature gets +1/+1
//	 until end of turn."
//
// Evasion is the shared RestrictSelf(CantBeBlocked) (Dimir Infiltrator); the
// pump is Deeproot Elite's "another Merfolk entered under your control"
// watcher with this creature's own until-end-of-turn boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "678095cc-ec8f-471a-b675-e59905d65c33",
		Name:         "River Sneak",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B10AnotherPermanentWithSubtypeEnteredUnderYourControl(ev, source, g, source.TextSubtype("Merfolk"))
			}, "River Sneak — this creature gets +1/+1 until end of turn", effects.ThisGetsUntilEndOfTurn(1, 1, "River Sneak — +1/+1")),
		},
	})
}
