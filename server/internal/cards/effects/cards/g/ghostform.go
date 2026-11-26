package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghostform — Sorcery {1}{U} (Core Set 2019, S82 OLD batch 200, #5527):
//
//	"Up to two target creatures can't be blocked this turn."
//
// Sparkmage's Gambit's up-to-two creature clause with Bessie's CantBeBlocked
// restriction in place of its can't-block: each target still legal as the
// spell resolves (CR 608.2b) can't be blocked until the cleanup step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d813b72-8ead-4a61-88ff-e93c8d843196",
		Name:         "Ghostform",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked, Label: "Ghostform — can't be blocked this turn"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
