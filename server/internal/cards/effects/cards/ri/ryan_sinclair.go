package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ryan Sinclair — Legendary Creature — Human {2}{R}, 2/2 (Doctor Who):
//
//	"Whenever Ryan attacks, exile cards from the top of your library until you
//	 exile a nonland card. You may cast the exiled card without paying its mana
//	 cost if it's a spell with mana value less than or equal to Ryan's power. Put the
//	 exiled cards not cast this way on the bottom of your library in a random order.
//	 Doctor's companion (You can have two commanders if the other is the Doctor.)"
//
// The run is the engine's exile-until-nonland (the lands go to exile with the nonland
// card). If that card's mana value is at most Ryan's power as the ability resolves
// (last known if he has gone; X is 0 in exile), a free cast is offered during the
// resolution (CR 608.2g); then, whatever was answered, every exiled card that was not
// cast goes to the bottom of the library in a random order. A card too expensive to
// cast skips the offer. Doctor's companion is a deck-construction rule read off the
// oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c06e2d5-1473-4ce6-bcf6-10dfbbe5f801",
		Name:         "Ryan Sinclair",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ryan Sinclair — exile until a nonland card; you may cast it free if its mana value is at most Ryan's power; put the rest on the bottom",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    effects.UntilNonland,
						ThenTail: ctx.ItemTail(ryanSinclairExiled),
					}.Apply(ctx)
				}),
		},
	})
}

// ryanSinclairExiled is the clause after the run, told the cards exiled. An on-disk key.
var ryanSinclairExiled = effects.CardTail("gap-b88/ryan-sinclair-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	g, controller := ctx.Game, ctx.Controller()
	hit := uuid.Nil
	for _, id := range o.IDs {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
			hit = id
		}
	}
	if hit != uuid.Nil {
		c, _ := g.LookupCardForEffect(hit)
		if z := g.FindCardZoneForEffect(hit); z != nil && z.Kind == game.ZoneExile && c.ManaValue() <= effects.B43PowerNowOrLastKnown(g, ctx.Source()) {
			perm := effects.ExileUntilNonlandFreeCastPermission(ctx, "Ryan Sinclair")
			perm.Player, perm.Zone = controller, game.ZoneExile
			if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
				Chooser:      controller,
				Source:       ctx.Source(),
				Reason:       "Ryan Sinclair",
				Cards:        []uuid.UUID{hit},
				Permission:   &perm,
				Keyword:      game.MayCastKeywordFreeCast,
				AcceptLabel:  "Cast it free",
				DeclineLabel: "Don't cast it",
				Then:         ryanSinclairRest,
				Carry:        o.IDs,
			}) {
				return nil
			}
		}
	}
	return gapB88BottomExiled(g, controller, o.IDs)
})

// ryanSinclairRest is "put the exiled cards not cast this way on the bottom", once the
// cast is answered: Carry is every card exiled. An on-disk key.
var ryanSinclairRest = game.RegisterChooseCardsThen("resolution-cast/ryan-sinclair-rest",
	func(g *game.Game, r game.CardsChosen) error {
		return gapB88BottomExiled(g, r.Chooser, r.Carry)
	})

// gapB88BottomExiled puts the cards that are still in exile on the bottom of the
// player's library in a random order; one that was cast has left exile and stays put.
func gapB88BottomExiled(g *game.Game, player uuid.UUID, ids []uuid.UUID) error {
	var rest []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			rest = append(rest, id)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneExile, rest)
}
