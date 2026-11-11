package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S39OpponentReturnsGreatest is one opponent's half of Dispersal:
// "returns a nonland permanent they control with the greatest mana value
// among permanents they control to its owner's hand, then discards a card".
//
// The greatest mana value is taken over EVERY permanent the player controls
// (lands included, at zero), and the candidates are the nonland ones that
// reach it; ties all qualify and the player picks. Mana value is read on the
// battlefield, where X is zero (CR 202.3e). The permanent is not targeted.
// One candidate is not asked; none means nothing returns, and the discard
// still happens.
func cmdr2S39OpponentReturnsGreatest(ctx *Context, player uuid.UUID) error {
	best := -1
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.ManaValue() > best {
			best = c.ManaValue()
		}
	}
	var candidates []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && !c.IsLand() && c.ManaValue() == best {
			candidates = append(candidates, c.InstanceID)
		}
	}
	switch len(candidates) {
	case 0:
		return cmdr2S39QueueDiscard(ctx.Game, player, ctx.Source())
	case 1:
		if err := ctx.Game.BounceToHandForEffect(candidates[0]); err != nil {
			return nothingIfGone(err)
		}
		return cmdr2S39QueueDiscard(ctx.Game, player, ctx.Source())
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   "Dispersal — return a nonland permanent with the greatest mana value to its owner's hand",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    cmdr2S39DispersalReturn,
		Asked:      game.NewAskedItem(ctx.Item),
	})
	return nil
}

func cmdr2S39QueueDiscard(g *game.Game, player, source uuid.UUID) error {
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: source, N: 1})
	return nil
}

// cmdr2S39DispersalReturn is the answer to Dispersal's pick: the chosen
// permanent returns to its owner's hand (if it is still on the battlefield
// under that player's control), then the same player discards a card.
// An on-disk key: never renamed, never reused.
var cmdr2S39DispersalReturn = game.RegisterChooseCardsThen("choose-cards/cmdr2-s39-dispersal-return", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if c, ok := g.LookupCardForEffect(id); !ok || c.Controller != r.Chooser {
			continue
		}
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		if err := nothingIfGone(g.BounceToHandForEffect(id)); err != nil {
			return err
		}
	}
	return cmdr2S39QueueDiscard(g, r.Chooser, r.Source)
})

// cmdr2S39JumpStart is "Jump-start" (CR 702.133a): cast this card from your
// graveyard by discarding a card in addition to paying its other costs, and
// exile it as it leaves the stack. It is Flashback's offer (the graveyard
// binding and the exile-on-leaving-the-stack replacement) at the card's own
// mana cost, plus Retrace's additional cost shape (a card from the hand,
// discarded as a cost, named in alt_cost_ids). Jump-start discards ANY card,
// not a land. The card file still opens the graveyard:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{cmdr2S39JumpStart("{1}{U}")},
func cmdr2S39JumpStart(printedCost string) game.AlternativeCost {
	ac := Flashback(printedCost)
	ac.Key = "jump-start"
	ac.Label = "Jump-start — discard a card"
	ac.DiscardFromHand = CardInYourHand("a card")
	ac.PayLabel = "a card"
	return ac
}
