package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dismember — Instant {1}{B/P}{B/P}:
//
//	"({B/P} can be paid with either {B} or 2 life.)
//	 Target creature gets -5/-5 until end of turn."
//
// Both Phyrexian symbols are paid by the engine at cast (CR 107.4f);
// the card file has only the -5/-5. Lethal toughness is the state
// check's business, as for every other shrink effect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd74f8eb-0253-42dd-8277-186d4934da38",
		Name:         "Dismember",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			return BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     -5,
				Toughness: -5,
				Label:     "Dismember — -5/-5",
			}.Apply(ctx)
		},
	})
}
