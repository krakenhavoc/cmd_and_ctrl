package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ill-Timed Explosion — Sorcery {2}{U}{R} (Murders at Karlov Manor,
// #4469):
//
//	"Draw two cards. Then you may discard two cards. When you do,
//	 Ill-Timed Explosion deals X damage to each creature, where X is the
//	 greatest mana value among cards discarded this way."
//
// Draw two, then a "you may" (MayChoice, nothing is asked on an empty
// hand), then the discard of two. Declining, or discarding nothing, ends
// the spell; having discarded, a reflexive trigger (ReflexiveTrigger,
// CR 603.12) carries X, the greatest mana value among the cards actually
// discarded (an {X} in a discarded card's cost is 0, the ruling), and
// deals that damage to each creature at once. The spell is the source,
// read as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6e3b3a3f-e32f-4c72-a9fb-dd1e877876b6",
		Name:         "Ill-Timed Explosion",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{N: 2}).Apply(ctx); err != nil {
				return err
			}
			if p := ctx.Game.PlayerByIDForEffect(item.Controller); p == nil || p.Hand.Size() == 0 {
				return nil
			}
			return MayChoice{
				Question: "Ill-Timed Explosion — discard two cards?",
				YesLabel: "Discard two",
				NoLabel:  "Decline",
				Then:     illTimedExplosionMay,
			}.Apply(ctx)
		},
	})
}

var illTimedExplosionMay = MayYes("confirm/ill-timed-explosion-discard", func(ctx *Context, _ game.ConfirmAnswer) error {
	return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player: ctx.Controller(),
		Source: ctx.Source(),
		N:      2,
	}, illTimedExplosionDiscarded, nil)
})

var illTimedExplosionDiscarded = game.RegisterDiscardThen("choose-cards/ill-timed-explosion-discard", func(g *game.Game, r game.DiscardsDone) error {
	discarded := r.Discarded.Cards()
	if len(discarded) == 0 {
		return nil
	}
	greatest := 0
	for _, id := range discarded {
		if c, ok := g.LookupCardForEffect(id); ok && c.ManaValue() > greatest {
			greatest = c.ManaValue()
		}
	}
	return ReflexiveTrigger{
		Label:  "Ill-Timed Explosion — damage to each creature equal to the greatest mana value among the discarded cards",
		Cards:  discarded,
		Body:   illTimedExplosionBody,
		Params: game.EffectParams{Amount: greatest},
	}.Apply(keyedContext(g, r.Player, r.Source))
})

var illTimedExplosionBody = game.DelayedBody("ill-timed-explosion/damage", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if p.Amount <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, c := range MatchingBattlefield(ctx, Creature()) {
		ids = append(ids, c.InstanceID)
	}
	return DealDamageToEachThen(ctx, ids, p.Amount, RecipientThen{})
})
