package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soul's Majesty — Sorcery {4}{G} (EDHREC rank 3312):
//
//	"Draw cards equal to the power of target creature you control."
//
// Green's card draw off a big body. The power is read at resolution
// — current power, so counters and anthems count — from the creature
// chosen at announce; a creature that left in response counters the
// spell (CR 608.2b) and nothing is drawn, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50fd4cf3-c347-4eb9-ab15-a9b0c5ea8b0f",
		Name:         "Soul's Majesty",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return DrawCards{Player: item.Controller, N: b31CurrentPowerOf(ctx.Game, t.ID)}.Apply(ctx)
			}
			return nil
		},
	})
}
