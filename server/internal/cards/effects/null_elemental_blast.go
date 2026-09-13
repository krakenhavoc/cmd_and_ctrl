package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Null Elemental Blast — Instant {C} (EDHREC rank 1575):
//
//	"Choose one —
//	 • Counter target multicolored spell.
//	 • Destroy target multicolored permanent."
//
// The colorless Blast. Two targeted options on a "choose one", so the
// chosen option's clause is the whole cast's — Rakdos Charm's shape.
// "Multicolored" is the effective colour count (two or more), read
// the same way for a spell on the stack and a permanent on the
// battlefield. The {C} cost is real: it wants a colorless mana, not
// a generic one, and the cost engine already enforces that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ec7699e3-ce56-46e4-9e4d-5ed2bd5fca83",
		Name:         "Null Elemental Blast",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Counter target multicolored spell.", TargetSpell("target multicolored spell", Multicolored())),
			Mode("Destroy target multicolored permanent.", TargetPermanent("target multicolored permanent", Multicolored())),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				switch {
				case ctx.HasMode(0):
					if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
						return err
					}
				case ctx.HasMode(1):
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
