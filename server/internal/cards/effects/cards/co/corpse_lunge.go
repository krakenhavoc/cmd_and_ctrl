package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corpse Lunge — {2}{B} Instant (Innistrad, S82 OLD b447, #5572):
//
//	"As an additional cost to cast this spell, exile a creature card from your
//	 graveyard.
//	 Corpse Lunge deals damage equal to the exiled card's power to target
//	 creature."
//
// Corpse Explosion's cost and read: the caster names the card as the spell is
// cast (CR 601.2b, 601.2h) and the power is read from the card where the cost
// left it, in exile (a power of 0 or less deals no damage, CR 120.8). The
// target is judged again as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1392cfbc-ea8a-4eb3-8680-ddd12a3bbb6d",
		Name:         "Corpse Lunge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		AdditionalCost: effects.ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool {
			return c.IsCreature()
		}),
		Targets: effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			power := 0
			for _, id := range ctx.Exiled() {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					power = c.CurrentPower()
				}
			}
			if power <= 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: power}.Apply(ctx)
				}
			}
			return nil
		},
	})
}
