package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alluring Scent — Sorcery, {1}{G}{G}:
//
//	"All creatures able to block target creature this turn do so."
//
// Taunting Challenge's text (#1684): a Lure record pinned to the
// target until end of turn (BlockRequirementUntilEOT), read by the
// engine exactly as Lure's static is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad218276-a44b-4a61-8e42-26a27929bbbb",
		Name:         "Alluring Scent",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return lureTargetForTheTurn(ctx, "Alluring Scent — all creatures able to block it do so")
		},
	})
}

// lureTargetForTheTurn is "All creatures able to block target creature
// this turn do so" (Alluring Scent, Bloodscent): a Lure record on the
// first still-legal target until end of turn. A target that left in
// response is skipped (CR 608.2b).
func lureTargetForTheTurn(ctx *Context, label string) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return BlockRequirementUntilEOT{Target: t.ID, Kind: game.BlockRequirementLure, Label: label}.Apply(ctx)
	}
	return nil
}
