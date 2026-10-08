package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reasonable Doubt — Instant {1}{U}:
//
//	"Counter target spell unless its controller pays {2}.
//	 Suspect up to one target creature. (A suspected creature has menace
//	 and can't block.)"
//
// Two target clauses (ADR 0065): the spell, then "up to one" creature,
// which may be left empty. Each is judged on its own at resolution
// (CR 608.2b), so a spell that was countered or left in response still
// lets the suspect happen, and a creature that left still lets the
// counter-unless be asked.
//
// The counter-unless is Dazzling Denial's: the SPELL's controller is
// asked, the stack entry is read before anything touches it, and a
// payment they cannot fund degrades to a decline. It only queues the
// question, so the suspect below it happens first — the two clauses
// touch different objects, so the order cannot be seen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "020503f5-0915-4cd5-a617-1c7bde045e10",
		Name:         "Reasonable Doubt",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetSpell("target spell"),
			UpToOneTargetCreature("up to one target creature"),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				if err := (Suspect{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			spell, ok := ctx.ClauseTarget(0)
			if !ok || ctx.Game.StackItemForEffect(spell.ID) == nil {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  spell.ID,
				Cost:     "{2}",
				Question: "Reasonable Doubt — pay {2} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
