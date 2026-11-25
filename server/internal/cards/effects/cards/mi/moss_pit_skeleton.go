package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moss-Pit Skeleton — Creature — Plant Skeleton {B}{G}, 2/2 (Zendikar
// Rising, S82 OLD #5508):
//
//	"Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	 If this creature was kicked, it enters with three +1/+1 counters on it.
//	 Whenever one or more +1/+1 counters are put on a creature you control,
//	 if this card is in your graveyard, you may put this card on top of your
//	 library."
//
// The entry counters are the kicked ones. The second trigger works from
// the graveyard (CR 113.6k), where "you" is the card's owner; it is an
// intervening "if" (CR 603.4) so the card must still be there as it
// resolves, and putting it on top of the library is the choice's yes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "edf0e7c0-d9a1-4d29-9cee-5267d3a3978f",
		Name:          "Moss-Pit Skeleton",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{3}")},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{
			effects.CountersPerKick(game.CounterPlusOne, 3),
		},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B39PlusOneCountersOnCreatureYouControl(ev, source, g)
				return ok
			}, "Moss-Pit Skeleton — you may put this card on top of your library",
				func(g *game.Game, item *game.StackItem) error {
					if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
						return nil
					}
					return effects.MayChoice{
						Question: "Moss-Pit Skeleton — put it on top of your library?",
						YesLabel: "Put it on top",
						NoLabel:  "Leave it",
						Then:     mossPitSkeletonAnswer,
					}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}

var mossPitSkeletonAnswer = effects.MayYes("confirm/moss-pit-skeleton-top-of-library", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if ctx.IsNewSourceObject(ctx.Source()) {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(ctx.Source()); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ctx.Game.TuckToLibraryForEffect(ctx.Source(), false)
})
