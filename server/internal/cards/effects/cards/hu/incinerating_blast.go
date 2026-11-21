package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incinerating Blast — Sorcery {4}{R} (Foundations, #4465):
//
//	"Incinerating Blast deals 6 damage to target creature.
//	 You may discard a card. If you do, draw a card."
//
// Joust Through's damage step at 6, then Flaring Cinder's rummage
// (b39MayDiscardThenDraw): the discard is optional, and a card is drawn
// only for a card discarded. Each is its own step, so the rummage waits for
// any prompt the damage asks (CR 608.2c). A target that is gone as the
// spell resolves fizzles the whole spell, rummage included (CR 608.2b).
//
// No simplification.
func init() {
	const rummage = "Incinerating Blast — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "49f0da4a-e046-4220-b90e-e105a5f2ef00",
		Name:         "Incinerating Blast",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 6)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, ref := range ctx.LegalTargets() {
						if err := (effects.DealDamage{Source: ctx.Item.SourceCardID, Target: ref.ID, Amount: 6}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
			effects.Instruction(effects.B39MayDiscardThenDraw(1, false, rummage, effects.DrawOneIfDiscarded)),
		),
	})
}
