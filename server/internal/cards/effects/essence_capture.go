package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Essence Capture — Instant {U}{U} (S82 reprint pass):
//
//	"Counter target creature spell. Put a +1/+1 counter on up to one target
//	 creature you control."
//
// Two target clauses: the creature spell, then up to one creature you
// control. Each is judged again as the spell resolves (CR 608.2b), and one
// going illegal leaves the other half to happen: a spell that has already
// resolved or been countered does not stop the counter, and a creature
// that has left does not stop the counter spell. A countered spell is
// countered before the counter is placed, in the printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f730bd9-2060-46b1-9208-0ac6562e8b2a",
		Name:         "Essence Capture",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetSpell("target creature spell", Creature()),
			TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if spell, ok := ctx.ClauseTarget(0); ok {
				if err := (CounterTarget{StackID: spell.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			for _, t := range ctx.ClauseTargets(1) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
					return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
