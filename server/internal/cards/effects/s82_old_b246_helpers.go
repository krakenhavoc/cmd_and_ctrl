package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b246_helpers.go — shared pieces of Amonkhet batch 246 (#5535).

// YouHaveAtMostCardsInHand is "you have N or fewer cards in hand" for the
// player `you`, read live from the hand's size (Thresher Lizard).
func YouHaveAtMostCardsInHand(n int) func(g *game.Game, you uuid.UUID) bool {
	return func(g *game.Game, you uuid.UUID) bool {
		p := g.PlayerByIDForEffect(you)
		return p != nil && p.Hand != nil && p.Hand.Size() <= n
	}
}

// SelfPumpWhileYouHaveAtMostCardsInHand is "This creature gets +p/+t as long
// as you have N or fewer cards in hand" (Thresher Lizard): a layer 7c
// modification (CR 613.4c) of the source alone, read on every recompute and
// invalidated by every hand change (DependsOnHandSize).
func SelfPumpWhileYouHaveAtMostCardsInHand(n, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Label:             "Gets " + SignedPT(power, toughness) + " as long as you have " + NumberWord(n) + " or fewer cards in hand.",
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7C_Modify,
		DependsOnHandSize: true,
		AppliesTo:         WhileYouHave(YouHaveAtMostCardsInHand(n), SelfCreature),
		Apply:             PumpBy(power, toughness),
	}
}

// MayRemoveAMinusOneCounterFromCarried is "you may remove a -1/-1 counter from
// it" for the creature the prompt carries (Shed Weakness): a yes removes one if
// it still has any.
var MayRemoveAMinusOneCounterFromCarried = MayYes("confirm/s82-old-b246-remove-a-minus-one-counter", func(ctx *Context, r game.ConfirmAnswer) error {
	id := CarryAt(r, 0)
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || c.Counters[game.CounterMinusOne] <= 0 {
		return nil
	}
	return AddCounter{Target: id, Kind: game.CounterMinusOne, N: -1}.Apply(ctx)
})

// SelfMinusOnePerCardInYourHand is "This creature gets -1/-1 for each card in
// your hand" (Grim Strider, Stingerback Terror): a layer 7c modification
// (CR 613.4c) of the source alone, read live from its controller's hand and
// recomputed on every change to it (DependsOnHandSize).
func SelfMinusOnePerCardInYourHand() game.StaticAbility {
	return game.StaticAbility{
		Label:             "Gets -1/-1 for each card in your hand.",
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7C_Modify,
		DependsOnHandSize: true,
		AppliesTo:         SelfCreature,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
				n = p.Hand.Size()
			}
			c.Power -= n
			c.Toughness -= n
		},
	}
}
