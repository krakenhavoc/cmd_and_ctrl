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
// "If they do, you incubate 2" is NOT implemented. CounterUnlessPaid has
// only a decline branch (OnDecline) and no paid branch, and adding one is
// an engine change (game.CounterUnlessPaidPrompt) outside this slice. The
// card ships without the incubate, which is weaker than printed: the
// payer escapes cleanly instead of paying and handing the caster an
// Incubator. The seam is an OnPaid on that prompt.
func init() {
	Register(Spec{
		OracleID:     "a24d3d60-2bd7-40cb-8002-b092943f61fe",
		Name:         "Assimilate Essence",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The incubate when the spell's controller pays {4} isn't implemented — paying just lets the spell resolve.",
		},
		Targets: TargetSpell("target creature or battle spell", Or(Creature(), Battle())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{4}",
				Question: "Assimilate Essence — pay {4} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
