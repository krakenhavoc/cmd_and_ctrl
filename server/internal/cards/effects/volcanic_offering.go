package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Volcanic Offering — Instant {4}{R}:
//
//	"Destroy target nonbasic land you don't control and target nonbasic
//	 land of an opponent's choice you don't control.
//	 Volcanic Offering deals 7 damage to target creature you don't
//	 control and 7 damage to target creature of an opponent's choice
//	 you don't control."
//
// Four instances of "target", so four clauses in printed order (CR
// 601.2c): the caster's land, an opponent's choice of land, the
// caster's creature, an opponent's choice of creature (#2055, ADR 0019's
// 2026-10-16 amendment). The rulings: the opponents are chosen as the
// spell is cast, because they choose its targets; the caster may name
// the same opponent for both or different ones, and neither is a target;
// a chosen opponent may choose the very land or creature the caster did.
// "You don't control" is the caster's, on all four. Each target is
// checked on its own as the spell resolves (CR 608.2b), and the two
// destructions and the two damages each happen together.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "05e0a10d-532a-451c-a54e-27e64969fe0e",
		Name:         "Volcanic Offering",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target nonbasic land you don't control", NonbasicLand(), Not(YouControl())),
			OfAnOpponentsChoice(TargetPermanent("target nonbasic land of an opponent's choice you don't control", NonbasicLand(), Not(YouControl()))),
			TargetCreature("target creature you don't control", Not(YouControl())),
			OfAnOpponentsChoice(TargetCreature("target creature of an opponent's choice you don't control", Not(YouControl()))),
		),
		Purpose: ForTargets(DamageToTarget(2, 7), DamageToTarget(3, 7)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// The two lands are destroyed together, and once when both
			// clauses named the same one; then the 7s are dealt, the two
			// as one damage event, each to a target still legal (CR
			// 608.2b).
			var lands []uuid.UUID
			for slot := 0; slot < 2; slot++ {
				if t, ok := ctx.ClauseTarget(slot); ok && t.Kind == game.TargetCard && !slices.Contains(lands, t.ID) {
					lands = append(lands, t.ID)
				}
			}
			if len(lands) == 0 {
				return volcanicOfferingDamage(ctx.Game, item)
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect(lands, ctx.ItemTail(volcanicOfferingDestroyed))
		},
	})
}

// volcanicOfferingDamage is the second sentence: 7 damage to each of the
// two creature targets still legal, dealt at once.
func volcanicOfferingDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return g.DamageInstanceForEffect(func() error {
		for slot := 2; slot < 4; slot++ {
			t, ok := ctx.ClauseTarget(slot)
			if !ok || t.Kind != game.TargetCard {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 7}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

// volcanicOfferingDestroyed deals the 7s once the lands' destruction has
// settled, keyed (#4253 slice 10a).
var volcanicOfferingDestroyed = CardTail("then/volcanic-offering-damage", func(ctx *Context, _ game.TailOutcome) error {
	return volcanicOfferingDamage(ctx.Game, ctx.Item)
})
