package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Frantic Firebolt — Instant {2}{R} (Wilds of Eldraine, #4471):
//
//	"Frantic Firebolt deals X damage to target creature, where X is 2 plus
//	 the number of cards in your graveyard that are instant cards, sorcery
//	 cards, and/or have an Adventure."
//
// X is read as the spell resolves, with Frantic Firebolt itself still on the
// stack and so not in the graveyard. Each card counts once, whether it is
// an instant, a sorcery, an Adventure card, or several at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1882cc8-c81a-4131-b639-1bd6e083f8ba",
		Name:         "Frantic Firebolt",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.DealDamage{
				Source: ctx.Source(),
				Target: id,
				Amount: 2 + effects.Std54InstantSorceryOrAdventureInGraveyard(ctx.Game, item.Controller),
			}.Apply(ctx)
		},
	})
}
