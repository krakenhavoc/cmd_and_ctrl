package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspire Awe — Instant {3}{G}:
//
//	"Prevent all combat damage that would be dealt this turn except
//	 combat damage that would be dealt by enchanted creatures and
//	 enchantment creatures. Scry 2."
//
// #2026's "except": every source's combat damage is prevented unless,
// as it would be dealt, the source is enchanted (an Aura is attached to
// it) or is an enchantment (the ruling: "check immediately before damage
// would be dealt"). Only creatures deal combat damage (CR 510.1), so an
// enchantment source of combat damage is an enchantment creature. The
// scry happens as the spell resolves (the ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9ca539cd-9876-4c13-b220-d553c17f2378",
		Name:         "Inspire Awe",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			shield := PreventDamageFromSource{CombatOnly: true, Protect: ShieldAnything, Filter: game.DamageSourceFilter{
				Except: []game.PermanentQuery{{Enchanted: true}, QueryTypes("enchantment")},
			}}
			if err := shield.Apply(ctx); err != nil {
				return err
			}
			return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
