package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acolyte's Reward — Instant {1}{W}:
//
//	"Prevent the next X damage that would be dealt to target creature this turn, where X is your devotion to white. If damage is prevented this way, Acolyte's Reward deals that much damage to any target. (Each {W} in the mana costs of permanents you control counts toward your devotion to white.)"
//
// ADR 0108 owner decision 2 (#1906): the charged shield (CR 615.7) on the
// first target, X fixed as it resolves (the ruling), whose CR 615.5
// additional effect deals the amount prevented to the second target —
// carried on the shield as its follow-up's recipient (game.ShieldFollowUp
// To), pinned to the object it is now. The rulings, in order: an illegal
// first target means no shield and so no damage; an illegal second target
// still leaves the shield, which then deals nothing; after resolution
// neither target's legality is checked again, only whether it is still
// there to be dealt damage. The damage is Acolyte's Reward's own, not
// combat damage, and nothing for damage that can't be prevented
// (CR 615.12).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "944689b2-f9f7-45c3-952c-54287b9824d5",
		Name:         "Acolyte's Reward",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			TargetAny(),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			shielded, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			shield := PreventNextDamage{
				Target: shielded.ID,
				Amount: devotionTo(ctx.Game, ctx.Controller(), "W"),
				Then:   dealThatMuchToTheChosenTargetBody,
				Label:  "Acolyte's Reward — prevent the next damage to it and deal that much",
			}
			if to, ok := ctx.ClauseTarget(1); ok {
				shield.To = to.ID
			}
			return shield.Apply(ctx)
		},
	})
}
