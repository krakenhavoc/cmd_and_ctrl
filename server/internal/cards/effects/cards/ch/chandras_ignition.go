package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chandra's Ignition — Sorcery {3}{R}{R} (EDHREC rank 477):
//
//	"Target creature you control deals damage equal to its power to
//	 each other creature and each opponent."
//
// A one-sided wrath and a burn spell in one, scaled by your biggest
// creature. The damage is dealt BY THE CREATURE, not by the spell —
// the DealDamage source is the target's instance ID — so a damage
// doubler, a prevention shield, deathtouch and lifelink on that
// creature all apply exactly as in paper. Its power is read when the
// spell resolves, so a pump in response scales the whole effect.
//
// The creature itself is not damaged ("each OTHER creature"); every
// other creature on the battlefield is, including the caster's own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f61680da-606e-4d16-b0a0-361aa5210901",
		Name:         "Chandra's Ignition",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			source := item.Targets[0].ID
			card, ok := ctx.Game.LookupCardForEffect(source)
			if !ok {
				return nil
			}
			power := card.CurrentPower()
			if power <= 0 {
				return nil
			}
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, id := range ctx.CreatureIDs() {
					if id == source {
						continue
					}
					if err := (effects.DealDamage{Source: source, Target: id, Amount: power}).Apply(ctx); err != nil {
						return err
					}
				}
				for _, opp := range ctx.Opponents() {
					if err := (effects.DealDamage{Source: source, Target: opp, Amount: power}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
