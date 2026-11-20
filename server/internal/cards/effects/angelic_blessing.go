package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angelic Blessing — Sorcery {2}{W} (Portal, #5630):
//
//	"Target creature gets +3/+3 and gains flying until end of turn. (It can't be blocked except by creatures with flying or reach.)"
//
// Mighty Leap's body with +3/+3.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d3758fca-0522-4b5a-a1cc-3b2b3ab299ba",
		Name:         "Angelic Blessing",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "flying")),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(3, 3, "Angelic Blessing — +3/+3 and flying", "flying")(ctx.Game, item)
		},
	})
}
