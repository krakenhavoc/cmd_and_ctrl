package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glorious Gale — Instant {1}{U}:
//
//	"Counter target creature spell. If it was a legendary spell, the
//	 Ring tempts you."
//
// "It was a legendary spell" is read off the spell before it is
// countered. A legendary spell that can't be countered still was a
// legendary spell, so the Ring still tempts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "858265c2-6415-4938-a6ec-734d53046100",
		Name:         "Glorious Gale",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature spell", Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				legendary := ok && c.IsLegendary()
				if err := (CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if legendary {
					return TheRingTemptsYou{}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
