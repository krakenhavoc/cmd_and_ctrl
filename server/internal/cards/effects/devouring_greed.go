package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Devouring Greed — Sorcery — Arcane {2}{B}{B}:
//
//	"As an additional cost to cast this spell, you may sacrifice any
//	 number of Spirits. Target player loses 2 life plus 2 life for each
//	 Spirit sacrificed this way. You gain that much life."
//
// The variable sacrifice (ADR 0100 §3), read back from the announcement
// (ctx.Sacrificed). "That much" is the life actually lost, so the gain
// waits for the loss to settle through the CR 614 window
// (ChangePlayerLifeThenForEffect, #793): a player whose life total
// can't change loses nothing, and you gain nothing.
//
// Arcane matters only to splice, which reads the type line.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "efbe96b5-1f58-4818-a346-ea554813adbc",
		Name:           "Devouring Greed",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAnyNumberCost("any number of Spirits", HasSubtype("Spirit")),
		Targets:        TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 || legal[0].Kind != game.TargetPlayer {
				return nil
			}
			src, me := ctx.Source(), item.Controller
			loss := 2 + 2*ctx.Sacrificed()
			return ctx.Game.ChangePlayerLifeThenForEffect(src, legal[0].ID, -loss, func(g *game.Game, applied int) error {
				if applied >= 0 {
					return nil
				}
				return g.ChangePlayerLifeForEffect(src, me, -applied)
			})
		},
	})
}
