package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rip Apart — Sorcery {R}{W} (EDHREC rank 2528):
//
//	"Choose one —
//	 • Rip Apart deals 3 damage to target creature or planeswalker.
//	 • Destroy target artifact or enchantment."
//
// Boros's two-mana answer to most things. Two targeted options on a
// "choose one", so the chosen option's clause is the whole cast's —
// Null Elemental Blast's shape. The damage is the spell's own, so a
// planeswalker loses loyalty and a creature carries the damage until
// the cleanup step; the destroy honours indestructible through the
// single-target verb.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbdbf18f-0180-4ade-a79e-1e644dd42d6f",
		Name:         "Rip Apart",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Rip Apart deals 3 damage to target creature or planeswalker.",
				effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker()))), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.Mode("Destroy target artifact or enchantment.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment()))),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, t := range ctx.LegalTargets() {
					switch {
					case ctx.HasMode(0):
						if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
							return err
						}
					case ctx.HasMode(1):
						if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			})
		},
	})
}
