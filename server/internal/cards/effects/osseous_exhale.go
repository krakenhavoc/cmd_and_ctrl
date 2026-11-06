package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Osseous Exhale — Instant {1}{W} (Tarkir: Dragonstorm, #4463):
//
//	"As an additional cost to cast this spell, you may behold a Dragon.
//	 (You may choose a Dragon you control or reveal a Dragon card from
//	 your hand.)
//	 Osseous Exhale deals 5 damage to target attacking or blocking
//	 creature. If a Dragon was beheld, you gain 2 life."
//
// The optional behold is OptionalBehold (CR 701.4a), Hyperion's Atomic
// Vision's: the caster who takes the offer names one Dragon they control
// or one Dragon card in hand on reveal_ids. "If a Dragon was beheld" is a
// fact about the announcement (CR 701.4b), read with ctx.WasBeheld(), so
// what happens to the Dragon afterwards does not matter. The life is
// gained after the damage, in printed order; if the target is illegal as
// the spell resolves, it does not resolve and no life is gained (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "8d84a706-080c-4fb2-b1b9-3dbb3294fb7b",
		Name:          "Osseous Exhale",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(DamageToTarget(0, 5)),
		OptionalCosts: []game.AdditionalCost{OptionalBehold("a", "Dragon")},
		Targets:       TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageToFirstTarget(5)(item, ctx); err != nil {
				return err
			}
			if !ctx.WasBeheld() {
				return nil
			}
			return GainLife{Amount: 2}.Apply(ctx)
		},
	})
}
