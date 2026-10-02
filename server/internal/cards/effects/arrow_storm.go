package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arrow Storm — Sorcery {3}{R}{R}:
//
//	"Arrow Storm deals 4 damage to any target.
//	 Raid — If you attacked this turn, instead Arrow Storm deals 5
//	 damage to that permanent or player and the damage can't be
//	 prevented."
//
// Raid is read as the spell resolves, for both halves: the amount and
// the spell's own "can't be prevented" (SpellRaid, ADR 0107 §5), so
// the stack chip lights up the moment the condition holds.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "a0ce2f5a-9c0b-4801-9105-72f84d8a4a1f",
		Name:                       "Arrow Storm",
		Completeness:               CompletenessFull,
		SpellDamageCantBePrevented: SpellRaid(),
		Targets:                    TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amount := 4
			if SpellRaid()(ctx.Game, item) {
				amount = 5
			}
			return damageToFirstTarget(amount)(item, ctx)
		},
	})
}
