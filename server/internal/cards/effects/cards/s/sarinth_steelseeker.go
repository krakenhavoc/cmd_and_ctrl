package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sarinth Steelseeker — Creature — Human Artificer Scout {1}{G}, 1/2 (The
// Brothers' War, #5482):
//
//	"Whenever an artifact you control enters, look at the top card of your
//	 library. If it's a land card, you may reveal it and put it into your
//	 hand. If you don't put the card into your hand, you may put it into
//	 your graveyard."
//
// The look is private to the controller. A land card is offered to the
// hand first (revealed as it goes); anything not put into the hand, whether
// because it is not a land or because the offer was declined, is then
// offered to the graveyard (a plain move, not a mill; Bison Whistle's
// answer). Declining both leaves it on top. A card that left the library
// while the prompts waited is a new object and is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dacf7f0d-9ce2-4e28-98a5-e7ec270a4c83",
		Name:         "Sarinth Steelseeker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.ArtifactEnteredUnderYourControl,
				"Sarinth Steelseeker — look at the top card of your library; if it's a land card, you may reveal it and put it into your hand; if you don't, you may put it into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
					if len(looked) == 0 {
						return nil
					}
					c, ok := g.LookupCardForEffect(looked[0])
					if !ok {
						return nil
					}
					if c.IsLand() {
						return effects.MayChoice{
							Player:   ctx.Controller(),
							Question: "Sarinth Steelseeker — reveal " + c.Name + " and put it into your hand?",
							YesLabel: "Put it into your hand",
							NoLabel:  "Don't",
							Then:     sarinthSteelseekerTakeLand,
							Carry:    []uuid.UUID{looked[0]},
						}.Apply(ctx)
					}
					return sarinthSteelseekerOfferGraveyard(ctx, looked[0])
				}),
		},
	})
}

var sarinthSteelseekerTakeLand = effects.MayYesNo("s82-bro-b07/sarinth-take-land",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		id := effects.CarryAt(r, 0)
		if !effects.BisonWhistleStillInLibrary(ctx.Game, id) {
			return nil
		}
		ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Sarinth Steelseeker — reveal the top card")
		return ctx.Game.BounceToHandForEffect(id)
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		return sarinthSteelseekerOfferGraveyard(ctx, effects.CarryAt(r, 0))
	})

func sarinthSteelseekerOfferGraveyard(ctx *effects.Context, id uuid.UUID) error {
	if !effects.BisonWhistleStillInLibrary(ctx.Game, id) {
		return nil
	}
	return effects.MayChoice{
		Player:   ctx.Controller(),
		Question: "Sarinth Steelseeker — put the top card of your library into your graveyard?",
		YesLabel: "Put it into your graveyard",
		NoLabel:  "Leave it on top",
		Then:     sarinthSteelseekerToGraveyard,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

var sarinthSteelseekerToGraveyard = effects.MayYes("s82-bro-b07/sarinth-to-graveyard", effects.PutCarriedLibraryCardIntoGraveyard)
