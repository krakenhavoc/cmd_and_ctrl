package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fell — Sorcery {1}{B} (EDHREC rank 2822):
//
//	"Destroy target creature."
//
// Murder at sorcery speed, one mana cheaper. Single-target removal,
// through the verb that honours indestructible.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f968470-ce00-43a3-95f5-2e3fe987be19",
		Name:         "Fell",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
