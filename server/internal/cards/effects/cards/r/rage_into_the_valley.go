package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rage into the Valley — Sorcery {2}{B} (The Hobbit, HOB):
//
//	"You draw a card and lose 1 life.
//	 Amass Goblins 2."
//
// Profane Insight's draw and life loss (life loss, not damage, CR 119.3),
// then the amass; each waits for any prompt the one before asked (a dredge
// offer; CR 608.2c, #4155).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c7641b6-3a6c-4441-a98c-3b5d61d91f76",
		Name:         "Rage into the Valley",
		Purpose:      game.Purpose{Amass: 2},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 1},
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Item.SourceCardID, ctx.Item.Controller, -1)
			}),
			effects.Amass{Subtype: "Goblin", N: 2},
		),
	})
}
