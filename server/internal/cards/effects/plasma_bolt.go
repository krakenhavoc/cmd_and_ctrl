package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plasma Bolt — Sorcery {R} (Edge of Eternities):
//
//	"Plasma Bolt deals 2 damage to any target.
//	 Void — Plasma Bolt deals 3 damage instead if a nonland permanent
//	 left the battlefield this turn or a spell was warped this turn."
//
// Void (void.go) is read as the spell resolves. The damage is dealt once,
// at the size void gave it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac5258c1-5b46-4892-a445-bfd56875f039",
		Name:         "Plasma Bolt",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			n := 2
			if Void(ctx.Game) {
				n = 3
			}
			return DealDamage{
				Source: ctx.Source(),
				Target: item.Targets[0].ID,
				Amount: n,
			}.Apply(ctx)
		},
	})
}
