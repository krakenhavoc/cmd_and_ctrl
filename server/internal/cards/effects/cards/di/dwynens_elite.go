package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwynen's Elite — Creature — Elf Warrior {1}{G}, 2/2 (Foundations,
// #4465):
//
//	"When this creature enters, if you control another Elf, create a 1/1
//	 green Elf Warrior creature token."
//
// An intervening "if" (CR 603.4), written as Human Torch writes his: the
// other Elf is looked for as the creature enters, so the ability doesn't
// trigger without one, and again as the ability resolves, so losing the
// other Elf in response means no token. "Another" excludes this creature
// by instance, and "Elf" is read through the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d4a7dd3-9258-4bd1-adc4-08f0205e196a",
		Name:         "Dwynen's Elite",
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control another Elf",
				Test: func(ctx *effects.Context) bool {
					return effects.MshB4ControlsAnotherHero(ctx.Game, ctx.Controller(), ctx.Source(), ctx.TextSubtype("Elf"))
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && effects.MshB4ControlsAnotherHero(g, source.Controller, source.InstanceID, source.TextSubtype("Elf"))
			}, "Dwynen's Elite — create a 1/1 green Elf Warrior", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Elf Warrior"), N: 1}))),
		},
	})
}
