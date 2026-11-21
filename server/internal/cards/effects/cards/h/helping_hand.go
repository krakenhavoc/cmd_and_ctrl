package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Helping Hand — Sorcery {W}:
//
//	"Return target creature card with mana value 3 or less from your
//	 graveyard to the battlefield tapped."
//
// The tapped entry is stamped on the CR 614 move itself rather than
// tapped a beat later, so nothing sees it enter untapped. The target
// is re-checked at resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "170a5dbc-6a43-4715-9501-178b1eb7b08c",
		Name:         "Helping Hand",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
			effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
