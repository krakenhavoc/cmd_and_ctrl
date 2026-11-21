package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cosmic Hunger — Instant {1}{G} (March of the Machine):
//
//	"Target creature you control deals damage equal to its power to
//	 another target creature, planeswalker, or battle."
//
// A one-sided bite (Contest of Claws without the discover). The second
// pick must differ from the first; the damage is dealt by the first
// creature, equal to its power as the spell resolves. If either target is
// illegal then, nothing is dealt (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe84ce0b-cd25-41ee-82c1-8ab03df5791b",
		Name:         "Cosmic Hunger",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.Distinct(effects.S82MomB12CreaturePlaneswalkerOrBattle("another target creature, planeswalker, or battle")),
		),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			return effects.EachDealsDamageEqualToItsPower(ctx, []game.TargetRef{biter}, victim.ID)
		},
	})
}
