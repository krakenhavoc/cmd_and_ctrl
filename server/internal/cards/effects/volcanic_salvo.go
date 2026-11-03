package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volcanic Salvo — Sorcery {10}{R}{R} (Secrets of Strixhaven Commander,
// #4834):
//
//	"This spell costs {X} less to cast, where X is the total power of
//	 creatures you control.
//	 Volcanic Salvo deals 6 damage to each of up to two target creatures
//	 and/or planeswalkers."
//
// The reduction is Ghalta, Primal Hunger's self cost modifier: generic mana
// only (CR 601.2f), never below {R}{R}, and a total power of zero or less
// reduces nothing. The spell deals 6 to each of up to two distinct targets
// at once; a target that has become illegal takes nothing and the other still
// does (CR 608.2b). With no target chosen the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c2fd3a3-96f2-4e84-85a9-77698c2f83aa",
		Name:         "Volcanic Salvo",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(TotalPowerOfCreaturesYouControl(),
				"This spell costs {X} less to cast, where X is the total power of creatures you control."),
		},
		Targets: TargetPermanent("up to two target creatures and/or planeswalkers",
			Or(Creature(), Planeswalker())).WithCount(0, 2),
		Purpose: ForTargets(DamageToTarget(0, 6)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				ids = append(ids, t.ID)
			}
			return DealDamageToEachThen(ctx, ids, 6, RecipientThen{})
		},
	})
}
