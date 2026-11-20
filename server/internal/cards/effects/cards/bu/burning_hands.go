package cardsbu

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Burning Hands — Instant {1}{R} (Adventures in the Forgotten Realms):
//
//	"Burning Hands deals 2 damage to target creature or planeswalker. If that
//	 permanent is green, Burning Hands deals 6 damage instead."
//
// The colour is read through the layered view as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "51a0d1c2-ff29-4ac9-84a2-ad4f566fee8d",
		Name:         "Burning Hands",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			amount := 2
			if c, ok := ctx.Game.LookupCardForEffect(id); ok && slices.Contains(c.EffectiveColors(), ctx.TextColor("G")) {
				amount = 6
			}
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: amount}.Apply(ctx)
		},
	})
}
