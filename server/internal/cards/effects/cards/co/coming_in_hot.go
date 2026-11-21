package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coming In Hot — Instant {R} (March of the Machine, S82 #5477):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn.
//	 Scry 1."
//
// Both pump and keyword end in the cleanup step (CR 514.2). The scry is part
// of the same resolution and happens even when the target is no longer legal
// only if the spell still resolves (CR 608.2b: with its one target illegal
// it does not resolve at all).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9666729d-3a06-43f9-a957-d5cc4d25b974",
		Name:         "Coming In Hot",
		Purpose:      game.Purpose{Scry: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike"))},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Coming In Hot — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Coming In Hot — first strike"}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return effects.Scry{N: 1}.Apply(ctx)
		},
	})
}
