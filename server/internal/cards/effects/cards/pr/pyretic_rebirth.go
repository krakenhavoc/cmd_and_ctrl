package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyretic Rebirth — Instant {2}{B}{R} (Modern Horizons 3, #5948):
//
//	"Return target artifact or creature card from your graveyard to your
//	 hand. Pyretic Rebirth deals damage equal to that card's mana value to
//	 up to one target creature or planeswalker."
//
// Two target clauses read positionally: slot 0 is the card, slot 1 the
// optional recipient. The damage is the card's mana value as it lay in the
// graveyard, read before it moves (a card that is no longer a legal target
// at resolution deals nothing, since "that card" is not returned, CR
// 608.2b). The recipient may be skipped, and then only the return happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a38ace5b-34f5-4105-b6ee-dedb86037729",
		Name:         "Pyretic Rebirth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // damages a creature or planeswalker
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target artifact or creature card from your graveyard",
				effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature())),
			effects.TargetPermanent("up to one target creature or planeswalker",
				effects.Or(effects.Creature(), effects.Planeswalker())).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			card, ok := ctx.ClauseTarget(0)
			if !ok || card.Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(card.ID)
			if !ok {
				return nil
			}
			mv := c.ManaValue()
			if err := (effects.ReturnFromGraveyard{Target: card.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
			victim, ok := ctx.ClauseTarget(1)
			if !ok || victim.Kind != game.TargetCard {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: victim.ID, Amount: mv}.Apply(ctx)
		},
	})
}
