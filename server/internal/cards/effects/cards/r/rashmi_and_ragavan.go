package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rashmi and Ragavan — Legendary Creature — Elf Monkey {1}{G}{U}{R}, 2/4:
//
//	"Whenever you cast your first spell during each of your turns, exile the
//	 top card of target opponent's library and create a Treasure token. Then
//	 you may cast the exiled card without paying its mana cost if it's a
//	 spell with mana value less than the number of artifacts you control. If
//	 you don't cast it this way, you may cast it this turn."
//
// The trigger is "your first spell on your own turn": the cast tally, bumped
// before the cast event fires, says it is the controller's first this turn
// and it is the controller's turn.
//
// The resolution runs in printed order. The exile is the keyed top-of-library
// exile; its tail makes the Treasure (so it counts as an artifact), reads the
// artifact count, and then:
//
//   - a spell whose mana value is below that count is offered as a free cast
//     during the resolution (CR 608.2g). Its decline, or its being refused,
//     falls to the next case through the offer's own continuation;
//   - otherwise, or on a decline, the exiled card gets an ordinary cast
//     permission for the rest of the turn, at its printed cost and its own
//     timing ("you may cast it this turn").
//
// A land can't be cast, so neither permission does anything for one. An empty
// library still makes the Treasure.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bdef246c-cf66-40a1-aae4-a591846e73ca",
		Name:         "Rashmi and Ragavan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return rashmiFirstSpellOnYourTurn(ev, source, g)
				}, "Rashmi and Ragavan — exile the top card of target opponent's library and create a Treasure token; you may cast the exiled card free if it's a spell with mana value less than the number of artifacts you control, otherwise you may cast it this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind == game.TargetPlayer {
								return g.ExileTopKeyedForEffect(t.ID, item.Controller, 1, ctx.ItemTail(rashmiExiled))
							}
						}
						return nil
					}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

// rashmiFirstSpellOnYourTurn is "your first spell during each of your turns".
func rashmiFirstSpellOnYourTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor != source.Controller {
		return false
	}
	active := g.Seats[g.Turn.ActiveSeat]
	if active == nil || active.ID != source.Controller {
		return false
	}
	return g.CastTallyFor(source.Controller).Total == 1
}

const rashmiName = "Rashmi and Ragavan"

var rashmiExiled = effects.CardTail("tail/s82-mom-b21-rashmi-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if err := (effects.CreateToken{Controller: ctx.Controller(), Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if len(o.IDs) != 1 {
		return nil
	}
	g, id := ctx.Game, o.IDs[0]
	card, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	artifacts := effects.CountControlled(g, ctx.Controller(), func(c game.Card) bool { return c.IsArtifact() })
	if !card.IsLand() && card.ManaValue() < artifacts {
		perm := g.FreeCastFromExilePermission(ctx.Controller(), ctx.Source(), rashmiName)
		offered := g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
			Chooser:      ctx.Controller(),
			Source:       ctx.Source(),
			Reason:       rashmiName,
			Cards:        []uuid.UUID{id},
			Permission:   &perm,
			Keyword:      game.MayCastKeywordFreeCast,
			AcceptLabel:  "Cast it free",
			DeclineLabel: "Don't cast it this way",
			Then:         rashmiAfterFreeOffer,
			Carry:        []uuid.UUID{id},
		})
		if offered {
			return nil
		}
	}
	rashmiGrantThisTurn(g, ctx.Controller(), ctx.Source(), id)
	return nil
})

// rashmiAfterFreeOffer is "if you don't cast it this way, you may cast it
// this turn": when the free offer was answered without a cast, the card
// still in exile gets the ordinary permission. An on-disk key.
var rashmiAfterFreeOffer = game.RegisterChooseCardsThen("resolution-cast/s82-mom-b21-rashmi-after", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) > 0 || len(r.Carry) == 0 {
		return nil
	}
	rashmiGrantThisTurn(g, r.Chooser, r.Source, r.Carry[0])
	return nil
})

// rashmiGrantThisTurn lets the player cast the exiled card this turn at its
// printed cost and with its own timing.
func rashmiGrantThisTurn(g *game.Game, player, source, id uuid.UUID) {
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return
	}
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player:     player,
		Zone:       game.ZoneExile,
		CastOnly:   true,
		Source:     source,
		SourceName: rashmiName,
		Label:      rashmiName + " — cast the exiled card this turn",
	})
}
