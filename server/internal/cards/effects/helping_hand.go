package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "170a5dbc-6a43-4715-9501-178b1eb7b08c",
		Name:         "Helping Hand",
		Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
			YouOwn(), Creature(), ManaValueLE(3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
