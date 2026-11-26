package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghitu Journeymage — Creature — Human Wizard {2}{R}, 3/2 (Dominaria, S82
// OLD, #5530):
//
//	"When this creature enters, if you control another Wizard, this creature
//	 deals 2 damage to each opponent."
//
// An intervening "if" (CR 603.4), checked as the Journeymage enters and again
// as the trigger resolves; Wizards are read through the layers and the
// Journeymage never counts itself ("another"). Resistance Squad's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5202c00a-5e47-43fe-9ed7-52c276120df3",
		Name:         "Ghitu Journeymage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control another Wizard",
				Test: func(ctx *effects.Context) bool {
					return effects.S82VowB02ControlsAnother(ctx.Game, ctx.Controller(), ctx.Source(), ctx.TextSubtype("Wizard"))
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, game.Characteristic{}, g) && effects.S82VowB02ControlsAnother(g, source.Controller, source.InstanceID, source.TextSubtype("Wizard"))
			}, "Ghitu Journeymage — 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2}))),
		},
	})
}
