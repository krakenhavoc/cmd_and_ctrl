package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b054_helpers.go — shared bodies for OLD batch 54 (#5504).

// OldB054ReturnCreatureThenChosen is Infernal Offering's last sentence:
// "Return a creature card from your graveyard to the battlefield, then
// that player returns a creature card from their graveyard to the
// battlefield." The controller picks first and the chosen player (the
// item's most recent ChoosePlayer answer) picks after the controller's
// card has entered, each from their own graveyard, each mandatory (a
// graveyard with no creature card is skipped, and the other player still
// returns theirs).
func OldB054ReturnCreatureThenChosen(ctx *Context) error {
	return OldB054OfferCreature(ctx, ctx.Controller(), true)
}

// OldB054OfferCreature asks `player` for one creature card of their own
// graveyard. With thenChosen the item's chosen player is asked after the
// pick has entered (through the keyed tail); without it the pick is the
// last step.
func OldB054OfferCreature(ctx *Context, player uuid.UUID, thenChosen bool) error {
	key := oldB054SecondPicked
	if thenChosen {
		key = oldB054FirstPicked
	}
	if !oldB054QueueCreaturePrompt(ctx, player, key) && thenChosen {
		return oldB054OfferChosen(ctx)
	}
	return nil
}

// oldB054OfferChosen is the chosen player's pick, the last step.
func oldB054OfferChosen(ctx *Context) error {
	oldB054QueueCreaturePrompt(ctx, ctx.ChosenPlayer(), oldB054SecondPicked)
	return nil
}

// oldB054QueueCreaturePrompt queues the mandatory pick of one creature
// card from `player`'s graveyard, answered by `then`. It reports false,
// and asks nothing, when there is none to pick.
func oldB054QueueCreaturePrompt(ctx *Context, player uuid.UUID, then game.ChooseCardsThen) bool {
	if player == uuid.Nil {
		return false
	}
	candidates := GraveyardCardIDs(ctx, player, func(c game.Card) bool {
		return c.IsCreature() && !c.IsToken()
	})
	if len(candidates) == 0 {
		return false
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:     player,
		FromPlayer:  player,
		Source:      ctx.Source(),
		Question:    "Infernal Offering — return a creature card from your graveyard to the battlefield",
		Cards:       candidates,
		Min:         1,
		Max:         1,
		Destination: game.ChooseOntoBattlefield,
		ThenKey:     then,
		LiveZone:    game.ZoneGraveyard,
		Asked:       game.NewAskedItem(ctx.Item),
	})
	return true
}

// oldB054FirstPicked is the controller's pick: the card enters, and the
// chosen player is asked from the entry's own continuation, after every
// entry question is answered. An on-disk key.
var oldB054FirstPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b054-first-picked",
	func(g *game.Game, r game.CardsChosen) error {
		ctx, _ := AskedContext(g, r.Asked, r.Chooser, r.Source, nil)
		if !r.Answered || len(r.Picked) == 0 {
			return oldB054OfferChosen(ctx)
		}
		return ReturnFromGraveyardTogether{Targets: r.Picked, ThenTail: oldB054ChosenReturns}.Apply(ctx)
	})

// oldB054SecondPicked is the last pick: the card enters. An on-disk key.
var oldB054SecondPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b054-second-picked",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		ctx, _ := AskedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return ReturnFromGraveyardTogether{Targets: r.Picked}.Apply(ctx)
	})

// oldB054ChosenReturns is "then that player returns a creature card from
// their graveyard", run once the first card has entered. An on-disk key.
var oldB054ChosenReturns = TogetherTail("together/s82-old-b054-chosen-returns",
	func(ctx *Context, _ []uuid.UUID, _ int) error {
		return oldB054OfferChosen(ctx)
	})

// OldB054PermanentTypes are the six permanent types (CR 110.4).
var OldB054PermanentTypes = []func(game.Card) bool{
	game.Card.IsArtifact, game.Card.IsBattle, game.Card.IsCreature,
	game.Card.IsEnchantment, game.Card.IsLand, game.Card.IsPlaneswalker,
}

// OldB054OnePerPermanentType is Revival Experiment's set rule: "for each
// permanent type, up to one card of that type". The picked cards are a
// legal set when each can be given a permanent type it has, no two the same
// (a bipartite matching of cards to types). A card of no permanent type
// cannot be picked. It is a function of the cards alone.
func OldB054OnePerPermanentType(picked []game.Card) bool {
	if len(picked) > len(OldB054PermanentTypes) {
		return false
	}
	owner := make([]int, len(OldB054PermanentTypes))
	for i := range owner {
		owner[i] = -1
	}
	var assign func(card int, seen []bool) bool
	assign = func(card int, seen []bool) bool {
		for t, has := range OldB054PermanentTypes {
			if seen[t] || !has(picked[card]) {
				continue
			}
			seen[t] = true
			if owner[t] == -1 || assign(owner[t], seen) {
				owner[t] = card
				return true
			}
		}
		return false
	}
	for c := range picked {
		if !assign(c, make([]bool, len(OldB054PermanentTypes))) {
			return false
		}
	}
	return true
}
