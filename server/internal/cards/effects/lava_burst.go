package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lava Burst — Sorcery {X}{R}:
//
//	"Lava Burst deals X damage to any target. If Lava Burst would deal
//	 damage to a creature, that damage can't be prevented or dealt
//	 instead to another permanent or player."
//
// The rider depends on the target, so it is the damage instruction's
// own mark (DealDamage.CantBePrevented and CantBeRedirected, ADR 0107
// §5) and only when the target is a creature as the damage is dealt.
// "Dealt instead" is a redirection, which the engine's redirection ban
// stops; no catalogued card redirects damage today.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f7be3da5-55b2-46f2-a5aa-277dee242b94",
		Name:         "Lava Burst",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			t := targets[0]
			creature := false
			if t.Kind == game.TargetCard {
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok {
					creature = c.IsCreature()
				}
			}
			return DealDamage{
				Source: item.SourceCardID, Target: t.ID, Amount: ctx.X(),
				CantBePrevented: creature, CantBeRedirected: creature,
			}.Apply(ctx)
		},
	})
}
