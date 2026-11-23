package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Orim's Thunder — Instant {2}{W} (Dominaria Remastered, #5948):
//
//	"Kicker {R}
//	 Destroy target artifact or enchantment. If this spell was kicked, it
//	 deals damage equal to that permanent's mana value to target creature."
//
// The creature target exists only when the spell is kicked (CR 601.2c). The
// permanent's mana value is read as it is destroyed (last-known
// information, CR 608.2h), before the damage is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "380429d5-82db-449c-b9b9-3e82ab987972",
		Name:         "Orim's Thunder",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Kicker("{R}"),
			effects.Clauses(
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				effects.TargetCreature("target creature"),
			))},
		Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				perm, ok := ctx.ClauseTarget(0)
				if !ok || perm.Kind != game.TargetCard {
					return nil
				}
				// The damage rides the tail as its N, and only a kicked spell has any.
				mv := 0
				if c, found := ctx.Game.LookupCardForEffect(perm.ID); found && ctx.WasKicked() {
					mv = c.ManaValue()
				}
				return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{perm.ID}, effects.WithN(ctx.ItemTail(orimsThunderDamage), mv))
			}),
		),
	})
}

// orimsThunderDamage is the kicked half: the spell deals damage equal to the
// destroyed permanent's mana value (read before it left, and zero unless the
// spell was kicked) to target creature.
// An on-disk key: never renamed, never reused.
var orimsThunderDamage = effects.CardTail("s82-gap-b142/orims-thunder-damage", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	creature, ok := ctx.ClauseTarget(1)
	if !ok || creature.Kind != game.TargetCard {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: creature.ID, Amount: o.Tail.N}.Apply(ctx)
})
