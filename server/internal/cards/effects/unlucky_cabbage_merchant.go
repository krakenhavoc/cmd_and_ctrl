package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unlucky Cabbage Merchant — Creature — Human Citizen {1}{G}, 2/2
// (Avatar: The Last Airbender, #4459):
//
//	"When this creature enters, create a Food token.
//	 Whenever you sacrifice a Food, you may search your library for a
//	 basic land card and put it onto the battlefield tapped. If you search
//	 your library this way, put this creature on the bottom of its owner's
//	 library, then shuffle."
//
// The sacrifice trigger is Rapacious Guest's (b31YouSacrificedAFood): any
// Food artifact, token or not (the ruling). The "may" is the trigger's
// own question (CR 603.5). Saying yes is searching, so the Merchant goes
// even when the search finds nothing (CR 701.23b): the land enters
// tapped, then the Merchant goes to the bottom of its owner's library if
// it is still the same object on the battlefield (CR 400.7), then the
// searched library is shuffled, the Merchant with it when it is yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8f7d4c8-bc50-4bcf-b930-797dcb3c2a21",
		Name:         "Unlucky Cabbage Merchant",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Unlucky Cabbage Merchant — create a Food", Do(CreateToken{Template: FoodToken(), N: 1})),
			Optional(On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b31YouSacrificedAFood(ev, source, g)
			}, "Unlucky Cabbage Merchant — search for a basic land, then put this creature on the bottom of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return SearchLibrary{
						Player:        ctx.Controller(),
						Predicate:     b30IsBasicLandCard,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						TappedOnEntry: true,
						Source:        ctx.Source(),
						Reason:        "Unlucky Cabbage Merchant — a basic land card to put onto the battlefield tapped",
						ThenTail:      ctx.ItemTail(unluckyCabbageMerchantTuck),
					}.Apply(ctx)
				}),
				"Unlucky Cabbage Merchant — search your library for a basic land card? (This creature goes to the bottom of its owner's library.)"),
		},
	})
}

// unluckyCabbageMerchantTuck is "put this creature on the bottom of its
// owner's library, then shuffle", after the search has put its land.
var unluckyCabbageMerchantTuck = CardTail("unlucky-cabbage-merchant/after-search", func(ctx *Context, _ game.TailOutcome) error {
	you := ctx.Controller()
	if ctx.Item == nil || !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return ctx.Game.ShuffleLibraryForEffect(you)
	}
	return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{ctx.Item.SourceCardID},
		game.TuckOptions{ToBottom: true}, game.KeyedTail{Ref: thenShuffle, Player: you})
})
