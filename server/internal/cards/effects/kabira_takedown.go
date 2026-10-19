package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kabira Takedown // Kabira Plateau — Modal DFC. Front: Instant {1}{W}:
// "Kabira Takedown deals damage equal to the number of creatures you
// control to target creature or planeswalker." Back: Land (mdfc_lands.go).
//
// The count is taken as the spell resolves (CR 608.2h), from the
// creatures the caster controls then. A target that is no longer legal
// means the spell does nothing. No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bb73c07-0220-4ba9-8d85-3c357c223833",
		Name:         "Kabira Takedown",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := len(MatchingBattlefield(ctx, And(Creature(), YouControl())))
			if n <= 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: id, Amount: n}.Apply(ctx)
		},
	})
}
