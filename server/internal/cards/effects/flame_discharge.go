package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flame Discharge — Instant {X}{R} (Kamigawa: Neon Dynasty):
//
//	"Flame Discharge deals X damage to target creature or planeswalker. If
//	 you controlled a modified creature as you cast this spell, it deals X
//	 plus 2 damage instead. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// "As you cast" fixes the answer when the spell is cast (CR 601.2i,
// Spec.CastTimeValues); a modified creature that dies in response still
// counts, and one that appears in response does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "fa4755c2-e573-45c3-bd6a-61b2b35fbd24",
		Name:           "Flame Discharge",
		Completeness:   CompletenessFull,
		XMatters:       true,
		Purpose:        game.Purpose{Targets: game.ForTargets(DamageXToTarget(0)), Answers: game.AnswerRemove},
		Targets:        TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		CastTimeValues: neoB10CastTimeModified(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			amount := ctx.X()
			if neoB10CastModified(ctx) > 0 {
				amount += 2
			}
			return DealDamageToTheTarget(amount)(ctx.Game, item)
		},
	})
}
