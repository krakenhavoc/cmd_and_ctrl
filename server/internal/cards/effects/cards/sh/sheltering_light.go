package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sheltering Light — Instant {W} (Ixalan, S82 older sets):
//
//	"Target creature gains indestructible until end of turn. Scry 1."
//
// Coming In Hot's shape: the keyword to the target if it is still legal, then
// the scry, which happens even when the target has gone as long as the spell
// resolves (CR 608.2b: one legal target keeps it resolving; with none it
// does not resolve at all).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "551ac58e-3a27-4ee3-ac91-16873562c0bc",
		Name:         "Sheltering Light",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible"))},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Sheltering Light — indestructible"}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return effects.Scry{N: 1}.Apply(ctx)
		},
	})
}
