package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sick and Tired — Instant {2}{B} (Urza's Legacy, S82 OLD b715, #5623):
//
//	"Two target creatures each get -1/-1 until end of turn."
//
// Exactly two targets, each still legal as the spell resolves gets the
// -1/-1 (CR 608.2b): one that left is skipped, the other still shrinks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "070d82db-0851-4f6d-8bf7-1a79c4c2cfee",
		Name:         "Sick and Tired",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -1, -1)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: -1, Toughness: -1, Label: "Sick and Tired — -1/-1"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
