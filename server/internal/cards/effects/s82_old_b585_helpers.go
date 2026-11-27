package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b585_helpers.go — Saviors of Kamigawa, batch 585 (#5601): the delayed
// discard of Ideas Unbound and the damage tally of Rushing-Tide Zubera.

// oldB585DiscardThreeBody is the delayed trigger "discard three cards at the
// beginning of the next end step": the controller of the spell chooses and
// discards three (CR 701.9, CR 603.7). A registered key (ADR 0041 phase 3).
var oldB585DiscardThreeBody = game.SimpleDelayedBody("s82-old-b585/discard-three-at-next-end-step", func(g *game.Game, item *game.StackItem) error {
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: item.Controller, Source: item.SourceCardID, N: 3})
	return nil
})

// OldB585DiscardThreeAtTheNextEndStep schedules the delayed discard.
func OldB585DiscardThreeAtTheNextEndStep(label string) Applier {
	return oldB585Step(func(ctx *Context) error {
		return ScheduleDelayedTrigger{At: game.StepEnd, Label: label, Body: oldB585DiscardThreeBody}.Apply(ctx)
	})
}

type oldB585Step func(ctx *Context) error

func (f oldB585Step) Apply(ctx *Context) error { return f(ctx) }

// OldB585DamageDealtToThisAtLeast is the intervening "if N or more damage was
// dealt to it this turn" of a dies trigger, read off the dying object.
func OldB585DamageDealtToThisAtLeast(n int) Condition {
	return CondBoard{Name: "damage dealt to it this turn", Test: func(ctx *Context) bool {
		return ctx.Game.TotalDamageDealtToObjectThisTurnForEffect(ctx.Item.SourceCardID) >= n
	}}
}

// OldB585PainsRewardWon is the high bidder's win: they lose life equal to the
// high bid and draw four cards (Pain's Reward). A high bidder who has left the
// game (r.Chooser is uuid.Nil) does neither.
var OldB585PainsRewardWon = NumberThen("number/old-b585-pains-reward-lose-the-bid-and-draw-four", func(ctx *Context, r game.NumberAnswer) error {
	if r.Chooser == uuid.Nil {
		return nil
	}
	if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -r.Number); err != nil {
		return err
	}
	return DrawCards{Player: r.Chooser, N: 4}.Apply(ctx)
})

// oldB585HandSize is the number of cards in `player`'s hand.
func oldB585HandSize(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Hand == nil {
		return 0
	}
	return len(p.Hand.Cards)
}

// OldB585RendingVinesResolve is Rending Vines: destroy the target if its mana
// value is at most the number of cards in the caster's hand, then draw a card.
func OldB585RendingVinesResolve(item *game.StackItem, ctx *Context) error {
	if id, ok := B16FirstLegalTargetCard(ctx); ok {
		if c, found := ctx.Game.LookupCardForEffect(id); found {
			if mv, ok := ctx.Game.ManaValueForEffect(c); ok && mv <= oldB585HandSize(ctx.Game, ctx.Controller()) {
				if err := (DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
		}
	}
	return DrawCards{N: 1}.Apply(ctx)
}

// OldB585DrawIfSevenCards is "Draw a card if you have seven or more cards in
// hand" (Scroll of Origins), the hand counted as the ability resolves.
func OldB585DrawIfSevenCards(g *game.Game, item *game.StackItem) error {
	if oldB585HandSize(g, item.Controller) < 7 {
		return nil
	}
	return DrawCards{N: 1}.Apply(NewContext(g, item))
}

// OldB585UnblockedAttacker matches an attacking creature no creature is
// blocking, its defending player's declaration complete (CR 509.1h).
func OldB585UnblockedAttacker() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		return g.UnblockedAttackerForEffect(c.InstanceID)
	}
}

// OldB585CurtainOfLightResolve makes the target unblocked attacker become
// blocked, then draws a card (Curtain of Light).
func OldB585CurtainOfLightResolve(item *game.StackItem, ctx *Context) error {
	if id, ok := B16FirstLegalTargetCard(ctx); ok {
		ctx.Game.BecomeBlockedForEffect([]uuid.UUID{id})
	}
	return DrawCards{N: 1}.Apply(ctx)
}
