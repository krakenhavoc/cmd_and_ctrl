package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aethermage's Touch — Instant {2}{W}{U} (S82 GAP, #5948):
//
//	"Reveal the top four cards of your library. You may put a creature card
//	 from among them onto the battlefield. It gains "At the beginning of your
//	 end step, return this creature to its owner's hand." Then put the rest of
//	 the cards revealed this way on the bottom of your library in any order."
//
// The reveal and the optional put are PutFromLibraryOntoBattlefield; its tail
// grants the creature the printed end-step trigger (a bundle granted with no
// duration, so it lasts while the object stays on the battlefield, CR 611.2a)
// and then orders the rest onto the bottom of the library in the player's
// chosen order. Fleeting Effigy's return-to-hand is the granted body.
func init() {
	const grantKey = "aethermages-touch/return-at-end-step"
	const text = "At the beginning of your end step, return this creature to its owner's hand."
	aethermagesTouchTail = effects.PutFromLibraryTail("put-from-library/aethermages-touch-grant-and-bottom",
		func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
			for _, id := range res.Entered {
				if !effects.OnBattlefield(ctx.Game, id) {
					continue
				}
				if err := (effects.GrantAbilitiesFor{
					Target:   id,
					Keys:     []string{grantKey},
					Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
					Label:    "Aethermage's Touch — returns to hand at the beginning of your end step",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return ctx.Game.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
				Chooser:   res.Player,
				Source:    res.Source,
				Cards:     res.Rest,
				From:      game.ZoneLibrary,
				Placement: game.LibraryPlaceBottom,
				Reason:    effects.LibraryOrderLabel(ctx.Game, res.Source, "put the rest on the bottom of your library in any order"),
			})
		})
	effects.Register(effects.Spec{
		OracleID:     "15692698-ef57-4672-bf76-5fe4a00c693a",
		Name:         "Aethermage's Touch",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Aethermage's Touch — return this creature to its owner's hand",
					effects.Do(effects.Instruction(effects.StdB9ReturnThisToHand))),
			},
			Text: text,
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 4, "Aethermage's Touch — reveal the top four cards"),
				Match:    effects.Creature(),
				Max:      1,
				Optional: true,
				Label:    "Aethermage's Touch — you may put a creature card from among them onto the battlefield",
				ThenTail: aethermagesTouchTail,
			}.Apply(ctx)
		},
	})
}

var aethermagesTouchTail game.TailRef
