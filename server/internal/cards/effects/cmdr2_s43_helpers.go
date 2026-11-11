package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s43_helpers.go — helpers of the S43 Commander-deck batch (impulse
// play, cast-from-top and life-for-mana costs). Names start cmdr2S43.

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

// cmdr2S43GrantedAltCostFreeHandOrTopOnceYourTurn is the on-disk key of One
// with the Multiverse's granted offer. Never renamed, never reused.
const cmdr2S43GrantedAltCostFreeHandOrTopOnceYourTurn = "granted-free-hand-or-top-once-your-turn"

// cmdr2S43CastFromHandOrTopFreeOnceEachOfYourTurns is One with the
// Multiverse's static: "Once during each of your turns, you may cast a spell
// from your hand or the top of your library without paying its mana cost."
// Omniscience's free offer (CastFromHandWithoutPayingManaCost), reaching the
// top of the library as well as the hand, once per turn per source and on
// your turn only. X is 0 (CR 107.3b).
func cmdr2S43CastFromHandOrTopFreeOnceEachOfYourTurns() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   cmdr2S43GrantedAltCostFreeHandOrTopOnceYourTurn,
			Label: "Cast it without paying its mana cost",
		},
		Zones:        []game.ZoneKind{game.ZoneHand, game.ZoneLibrary},
		OnceEachTurn: true,
		YourTurnOnly: true,
	}
}

// cmdr2S43TectonicGiantExile is "exile the top two cards of your library.
// Choose one of them. Until the end of your next turn, you may play that
// card" (Tectonic Giant's second mode): Strongbox Raider's shape
// (stdB15ExileTopTwoChooseOnePlay) worded for the Giant. Both cards are
// exiled face up, the controller picks one, and only it gets the window.
func cmdr2S43TectonicGiantExile(ctx *Context) error {
	ids := ctx.Game.LookAtTopOfLibraryForEffect(ctx.Controller(), 2)
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(cmdr2S43TectonicGiantChoose))
}

// cmdr2S43TectonicGiantChoose is "choose one of them" over the cards that
// reached exile. An on-disk key.
var cmdr2S43TectonicGiantChoose = CardTail("then/cmdr2-s43-tectonic-giant-choose", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Tectonic Giant — choose one of them (you may play it until the end of your next turn)",
		Cards:    o.IDs,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  cmdr2S43TectonicGiantPlay,
	})
	return nil
})

// cmdr2S43TectonicGiantPlay gives the chosen card its window. An on-disk key.
var cmdr2S43TectonicGiantPlay = game.RegisterChooseCardsThen("choose-cards/cmdr2-s43-tectonic-giant-play", func(g *game.Game, r game.CardsChosen) error {
	for _, id := range r.Picked {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player:   r.Chooser,
			Zone:     game.ZoneExile,
			Duration: g.UntilEndOfYourNextTurnDuration(r.Chooser),
			Label:    "Tectonic Giant — play it until the end of your next turn",
		})
	}
	return nil
})

// cmdr2S43TectonicGiantTrigger is "whenever this creature attacks or becomes
// the target of a spell an opponent controls": the source was declared as an
// attacker, or became the target of a SPELL (not an ability) whose controller
// is an opponent (EventBecomesTarget names the targeter's controller as Actor
// and its kind as TargeterKind).
func cmdr2S43TectonicGiantTrigger(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventAttack:
		return ThisAttacked(ev, source, game.Characteristic{}, g)
	case game.EventBecomesTarget:
		return ev.CardID == source.InstanceID && ev.TargeterKind == game.StackItemSpell &&
			ev.Actor != uuid.Nil && ev.Actor != source.Controller
	}
	return false
}
