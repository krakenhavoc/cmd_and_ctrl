package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Centaur Vinecrasher — Creature — Plant Centaur {3}{G}, 1/1 (Foundations
// Commander):
//
//	"Trample
//	 This creature enters with a number of +1/+1 counters on it equal to
//	 the number of land cards in all graveyards.
//	 Whenever a land card is put into a graveyard from anywhere, you may
//	 pay {G}{G}. If you do, return this card from your graveyard to your
//	 hand."
//
// The entry counters are Diregraf Colossus's counted "enters with" clause
// over every graveyard. The trigger watches from the graveyard (Bloodghast's
// shape, its controller the card's owner): one per land card put there, by
// whatever route — a death or sacrifice, a mill, a discard, a spell's
// resolution — through the one event each route emits
// (cmdrB9LandCardPutIntoAGraveyard). A land that arrives with the Vinecrasher
// in the same event still triggers it, because the harvest finds the card
// where it landed (the ruling). "You may pay" is the payment prompt of
// Conspiracy Theorist; the return is Blessed Ghoul's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "90629e9c-83c0-4eca-b4ed-d0494ded6704",
		Name:            "Centaur Vinecrasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterPlusOne, func(g *game.Game, _ *game.Card) int {
				return cmdrB9LandCardsInAllGraveyards(g)
			}, "Centaur Vinecrasher: enters with a +1/+1 counter for each land card in all graveyards"),
		},
		Triggered: []game.TriggeredAbility{
			InGraveyard(OnAny([]game.EventKind{game.EventZoneMove, game.EventMill, game.EventDiscardCard},
				cmdrB9LandCardPutIntoAGraveyard,
				"Centaur Vinecrasher — you may pay {G}{G} to return it from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{G}{G}",
						Question: "Centaur Vinecrasher — pay {G}{G} to return it to your hand?",
						Then:     centaurVinecrasherPaid,
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// centaurVinecrasherPaid is "if you do, return this card from your
// graveyard to your hand". An on-disk key.
var centaurVinecrasherPaid = PayUnlessBranches("pay-unless/centaur-vinecrasher-return",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return returnThisCardFromYourGraveyardToYourHand(ctx.Game, ctx.Item)
	}, nil)

// cmdrB9LandCardsInAllGraveyards counts the land cards in every graveyard.
func cmdrB9LandCardsInAllGraveyards(g *game.Game) int {
	n := 0
	for _, s := range g.Seats {
		n += b17LandCardsInGraveyard(g, s.ID)
	}
	return n
}

// cmdrB9LandCardPutIntoAGraveyard is "a land card is put into a graveyard
// from anywhere": the move's destination is a graveyard it did not come
// from, and the card is a land card (never a token, which is not a card).
// The source's own arrival is not excluded: a land card is the only thing
// that counts, and the Vinecrasher is not one.
func cmdrB9LandCardPutIntoAGraveyard(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsLand() && !c.IsToken()
}
