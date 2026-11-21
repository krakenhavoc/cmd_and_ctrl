package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infiltrate — Instant {U}
// (Nemesis, #5618):
//
//	"Target creature can't be blocked this turn."
//
// Artful Dodge without flashback: the shared
// targetCantBeBlockedThisTurn body, ending in the cleanup step (CR
// 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "081764b4-ef96-44aa-836c-05353efd215c",
		Name:         "Infiltrate",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetCantBeBlockedThisTurn("Infiltrate")(ctx.Game, item)
		},
	})
}
