package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Éomer of the Riddermark — Legendary Creature — Human Knight {4}{R}, 5/4
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Haste
//	 Whenever Éomer attacks, if you control a creature with the greatest
//	 power among creatures on the battlefield, create a 1/1 white Human
//	 Soldier creature token."
//
// High Score's greatest-power check (a creature tied for the greatest
// counts), as an intervening "if" on a self-attack watch: read as the
// attack is declared and again as the trigger resolves (CR 603.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c07340bb-04b6-4a80-9999-727b8ab3534d",
		Name:            "Éomer of the Riddermark",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control a creature with the greatest power among creatures on the battlefield",
				Test: func(ctx *effects.Context) bool {
					return effects.ControlsGreatestPowerAmongCreatures(ctx.Game, ctx.Controller())
				},
			}, effects.On(game.EventAttack, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ThisAttacked(ev, source, lki, g) && effects.ControlsGreatestPowerAmongCreatures(g, source.Controller)
			}, "Éomer of the Riddermark — create a 1/1 white Human Soldier token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1}))),
		},
	})
}
