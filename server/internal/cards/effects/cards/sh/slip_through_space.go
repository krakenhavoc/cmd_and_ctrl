package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slip Through Space — Sorcery {U} (Oath of the Gatewatch, #5545):
//
//	"Devoid (This card has no color.)
//	 Target creature can't be blocked this turn.
//	 Draw a card."
//
// Devoid makes it colourless (CR 702.114a). Enter the Enigma's text: the
// restriction lasts until cleanup, then the caster draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2516949d-75bb-4aea-a41d-0298bb129471",
		Name:            "Slip Through Space",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Draws: 1, Targets: game.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0))},
		Targets:         effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.TargetCantBeBlockedThisTurn("Slip Through Space")(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
