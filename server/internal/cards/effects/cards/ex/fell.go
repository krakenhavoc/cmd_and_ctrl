package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fell — Sorcery {1}{B} (EDHREC rank 2822):
//
//	"Destroy target creature."
//
// Murder at sorcery speed, one mana cheaper. Single-target removal,
// through the verb that honours indestructible.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f968470-ce00-43a3-95f5-2e3fe987be19",
		Name:         "Fell",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
