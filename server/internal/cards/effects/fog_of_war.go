package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fog of War — Instant {2}{G}:
//
//	"You gain 1 life for each creature on the battlefield. Prevent all
//	 combat damage that would be dealt this turn by creatures with power
//	 3 or less."
//
// The life counts every creature on the battlefield as the spell
// resolves (CR 608.2h). The shield is #2026's power bound, read as each
// creature would deal combat damage (CR 609.7b): a creature pumped past
// 3 after Fog of War resolved deals its damage, and one shrunk to 3 or
// less does not. Power counts its counters.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "103ca069-0bfe-4976-bee1-75406273875d",
		Name:         "Fog of War",
		Completeness: CompletenessFull,
		// The prevention waits for a prompt the life gain asks (a CR 616 order; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return GainLife{Player: ctx.Controller(), Amount: len(ctx.CreatureIDs())}.Apply(ctx)
			}),
			combatShieldAgainstCreatures(game.DamageSourceFilter{PowerBounded: true, PowerAtMost: 3}),
		),
	})
}
