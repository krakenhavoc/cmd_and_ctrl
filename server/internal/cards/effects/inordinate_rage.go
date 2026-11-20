package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inordinate Rage — Instant {1}{R} (Zendikar Rising, #5508):
//
//	"Target creature gets +3/+2 until end of turn. Scry 1."
//
// Titan's Strength's shape with +3/+2: the pump, then the scry (CR
// 701.22a). A spell whose only target became illegal is countered on
// resolution and does neither (CR 608.2b); one that resolves always
// scries.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50402103-14ea-49b0-8e66-d2b7f7a69a41",
		Name:         "Inordinate Rage",
		Purpose:      WithRiders(ForTargets(BenefitUntilEOT(0, 3, 2)), game.Purpose{Scry: 1}),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if id, ok := b16FirstLegalTargetCard(ctx); ok {
				if err := (BoostUntilEOT{Target: id, Power: 3, Toughness: 2,
					Label: "Inordinate Rage — +3/+2 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			return Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
