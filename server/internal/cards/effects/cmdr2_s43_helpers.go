package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s43_helpers.go — helpers of the S43 Commander-deck batch (impulse
// play, cast-from-top and life-for-mana costs). Names start cmdr2S43.

// cmdr2S43TargetOpponent is the player a "target opponent" ability resolved
// against, or uuid.Nil when the target is gone (CR 608.2b).
func cmdr2S43TargetOpponent(ctx *Context) uuid.UUID {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return t.ID
		}
	}
	return uuid.Nil
}

// cmdr2S43ExileUntilInstantOrSorcery is "target opponent exiles cards from
// the top of their library until they exile an instant or sorcery card. You
// may cast that card without paying its mana cost. Then put the exiled cards
// that weren't cast this way on the bottom of that library in a random
// order": the walk is the victim's, the free cast is the item controller's.
func cmdr2S43ExileUntilInstantOrSorcery(g *game.Game, item *game.StackItem, victim uuid.UUID) error {
	if victim == uuid.Nil {
		return nil
	}
	return g.ExileUntilCastFreeByForEffect(victim, item.Controller, item.SourceCardID,
		func(c game.Card) bool { return !c.IsToken() && (c.IsInstant() || c.IsSorcery()) })
}

// cmdr2S43StolenGoodsGranted is Stolen Goods' tail: "until end of turn, you
// may cast that card without paying its mana cost". The card stays in exile,
// the cast obeys the card's own timing (the card is not cast during the
// resolution), and a land is never offered.
var cmdr2S43StolenGoodsGranted = CardTail("cmdr2-s43/stolen-goods-grant", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		z := ctx.Game.FindCardZoneForEffect(id)
		if !ok || c.IsLand() || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     ctx.Controller(),
			Zone:       game.ZoneExile,
			Cost:       "{0}",
			CastOnly:   true,
			Duration:   ctx.Game.UntilEndOfTurnDuration(),
			Source:     ctx.Source(),
			SourceName: "Stolen Goods",
			Label:      "Cast without paying its mana cost (Stolen Goods)",
		}, []game.Card{c})
	}
	return nil
})

// cmdr2S43EpicExperimentOffer is Epic Experiment's tail, told the cards that
// landed in exile: it offers the free casts over them during the resolution
// (CR 608.2g) and, once that is answered, puts every card still in exile into
// the graveyard. An on-disk key.
var cmdr2S43EpicExperimentOffer = game.RegisterTail("cmdr2-s43/epic-experiment-offer", func(g *game.Game, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      o.Tail.Player,
		Source:       o.Tail.Source,
		Reason:       o.Tail.Word,
		Cards:        o.IDs,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: "Put them into your graveyard",
		Then:         cmdr2S43EpicExperimentRest,
		Carry:        o.IDs,
	}) {
		return nil
	}
	return cmdr2S43ExiledToGraveyard(g, o.IDs)
})

// cmdr2S43EpicExperimentRest is "put all cards exiled this way that weren't
// cast into your graveyard", once the casts are answered: Carry is every card
// the exile landed. An on-disk key.
var cmdr2S43EpicExperimentRest = game.RegisterChooseCardsThen("resolution-cast/cmdr2-s43-epic-experiment-rest",
	func(g *game.Game, r game.CardsChosen) error {
		return cmdr2S43ExiledToGraveyard(g, r.Carry)
	})

// cmdr2S43ExiledToGraveyard puts the members of ids still in exile into
// their owners' graveyards.
func cmdr2S43ExiledToGraveyard(g *game.Game, ids []uuid.UUID) error {
	var left []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			left = append(left, id)
		}
	}
	if len(left) == 0 {
		return nil
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(left, game.KeyedTail{})
}

// cmdr2S43GrantedAltCostLifeEnchantmentOnceYourTurn is the on-disk key of
// Demon of Fate's Design's granted offer. Never renamed, never reused.
const cmdr2S43GrantedAltCostLifeEnchantmentOnceYourTurn = "granted-life-enchantment-once-your-turn"

// cmdr2S43PayLifeForAnEnchantmentSpellOnceEachOfYourTurns is Demon of Fate's
// Design's static: "Once during each of your turns, you may cast an
// enchantment spell by paying life equal to its mana value rather than paying
// its mana cost." Access Maze's offer (PayLifeForASpellFromHandOnceEachOfYourTurns)
// narrowed to enchantment spells and to every zone a spell can be cast from
// (CR 118.9a).
func cmdr2S43PayLifeForAnEnchantmentSpellOnceEachOfYourTurns() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   cmdr2S43GrantedAltCostLifeEnchantmentOnceYourTurn,
			Label: "Pay life equal to its mana value rather than pay its mana cost",
		},
		Spells:               game.PermissionFilter{CardType: "enchantment"},
		OnceEachTurn:         true,
		YourTurnOnly:         true,
		LifeEqualToManaValue: true,
	}
}
