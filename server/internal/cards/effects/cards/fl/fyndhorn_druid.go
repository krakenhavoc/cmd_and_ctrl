package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fyndhorn Druid — Creature — Elf Druid {2}{G} 2/2 (Alliances, S82 OLD batch
// 820):
//
//	"When this creature dies, if it was blocked this turn, you gain 4 life."
//
// An intervening "if" (CR 603.4) about the creature as it was: any block this
// turn that named it as the attacker, whether or not it is still there to be
// blocked now.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98e51f1a-ed6e-4286-a41b-e4656d08c934",
		Name:         "Fyndhorn Druid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "it was blocked this turn",
				Test: func(ctx *effects.Context) bool { return effects.OldB820WasBlockedThisTurn(ctx.Game, ctx.Source()) },
			}, effects.WhenThisDies("Fyndhorn Druid — you gain 4 life", effects.Do(effects.GainLife{Amount: 4}))),
		},
	})
}
