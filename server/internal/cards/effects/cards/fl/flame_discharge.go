package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:       "fa4755c2-e573-45c3-bd6a-61b2b35fbd24",
		Name:           "Flame Discharge",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		Purpose:        game.Purpose{Targets: game.ForTargets(effects.DamageXToTarget(0)), Answers: game.AnswerRemove},
		Targets:        effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		CastTimeValues: effects.NeoB10CastTimeModified(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			amount := ctx.X()
			if effects.NeoB10CastModified(ctx) > 0 {
				amount += 2
			}
			return effects.DealDamageToTheTarget(amount)(ctx.Game, item)
		},
	})
}
