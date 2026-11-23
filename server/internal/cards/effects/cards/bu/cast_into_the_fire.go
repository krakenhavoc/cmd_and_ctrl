package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cast into the Fire — Instant {1}{R} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Choose one —
//	 • Cast into the Fire deals 1 damage to each of up to two target
//	   creatures.
//	 • Exile target artifact."
//
// A modal spell whose bullets target separately, each read from its own target
// group (CR 700.2). The damage is one event over every still-legal target
// (Ribbons of Night's DamageInstanceForEffect), the spell being the source.
// The exile is a plain exile of the artifact as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b24ed296-e17d-4e36-8a86-a370868b0136",
		Name:         "Cast into the Fire",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Cast into the Fire deals 1 damage to each of up to two target creatures.",
				effects.TargetCreature("each of up to two target creatures").WithCount(0, 2),
				func(item *game.StackItem, ctx *effects.Context, occurrence int) error {
					var ids []game.TargetRef
					for _, t := range ctx.ModeTargets(occurrence) {
						if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
							ids = append(ids, t)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					return ctx.Game.DamageInstanceForEffect(func() error {
						for _, t := range ids {
							if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					})
				}),
			effects.ModeDoing("Exile target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact()),
				func(_ *game.StackItem, ctx *effects.Context, occurrence int) error {
					for _, t := range ctx.ModeTargets(occurrence) {
						if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
							return effects.ExileTarget{Target: t.ID}.Apply(ctx)
						}
					}
					return nil
				}),
		),
	})
}
