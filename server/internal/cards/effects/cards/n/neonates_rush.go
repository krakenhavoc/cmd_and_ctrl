package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Neonate's Rush — Instant {2}{R} (Innistrad: Midnight Hunt, #5497):
//
//	"This spell costs {1} less to cast if you control a Vampire.
//	 Neonate's Rush deals 1 damage to target creature and 1 damage to its
//	 controller. Draw a card."
//
// Bolt Bend's self cost modifier, skipping the spell itself. The damage to
// the creature's controller is read from the creature as it is as the spell
// resolves; if the target is illegal the spell does not resolve at all (CR
// 608.2b) and nothing is drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "721aeedc-1de5-40d0-a04c-a2ebc056d06a",
		Name:         "Neonate's Rush",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 1)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you control a Vampire.", youControlAVampireOtherThanTheSpell),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			creature, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			owner := creature.Controller
			if err := (effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: owner, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}

// youControlAVampireOtherThanTheSpell is Neonate's Rush's discount
// condition: a Vampire the caster controls on the battlefield.
func youControlAVampireOtherThanTheSpell(q game.CostQuery) bool {
	if q.Game == nil || q.Game.Battlefield == nil {
		return false
	}
	for _, c := range q.Game.Battlefield.Cards {
		if c.Controller == q.Controller && c.HasSubtype(q.Source.TextSubtype("Vampire")) {
			return true
		}
	}
	return false
}
