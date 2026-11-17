package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fe84ce0b-cd25-41ee-82c1-8ab03df5791b",
		Name:         "Cosmic Hunger",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(s82MomB12CreaturePlaneswalkerOrBattle("another target creature, planeswalker, or battle")),
		),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			return eachDealsDamageEqualToItsPower(ctx, []game.TargetRef{biter}, victim.ID)
		},
	})
}
