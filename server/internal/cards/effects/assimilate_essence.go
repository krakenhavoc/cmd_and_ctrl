package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assimilate Essence — Instant {1}{U} (March of the Machine):
//
//	"Counter target creature or battle spell unless its controller pays
//	 {4}. If they do, you incubate 2."
//
// The counter-unless-pays half is CounterUnlessPaid (#951): the prompt
// goes to the spell's controller and holds the stack while it is open.
//
// "If they do, you incubate 2" is its OnPaid branch (#3591): it runs
// only after the {4} was actually paid, with the caster as "you" (the
// context is bound to this spell, not to the payer), and creates the
// Incubator once. A payer who says "Pay" and cannot fund it has
// declined: their spell is countered and nothing is incubated. If the
// spell has left the stack by the time the answer arrives, neither
// branch runs (the engine drops both), which is right: the "unless"
// bought off nothing.
func init() {
	Register(Spec{
		OracleID:     "a24d3d60-2bd7-40cb-8002-b092943f61fe",
		Name:         "Assimilate Essence",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target creature or battle spell", Or(Creature(), Battle())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{4}",
				Question: "Assimilate Essence — pay {4} or your spell is countered",
				OnPaid:   Incubate{N: 2}.Apply,
			}.Apply(ctx)
		},
	})
}
