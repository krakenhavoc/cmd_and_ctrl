package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unexpected Results — Sorcery {2}{G}{U} (Gatecrash, S82 OLD #5565):
//
//	"Shuffle your library, then reveal the top card. If it's a nonland card,
//	 you may cast it without paying its mana cost. If it's a land card, you
//	 may put it onto the battlefield and return Unexpected Results to its
//	 owner's hand."
//
// The shuffle, then a public reveal of the top card (Yennett's). A nonland card
// is offered as a free cast made during the resolution (CR 608.2g), from the
// library, through the resolution_cast prompt; declining leaves it on top. A land
// card is offered to be put onto the battlefield (a put, not a land play:
// CR 305.4, no land drop is used), and only if it really enters is the spell
// marked to go to its owner's hand instead of its graveyard as its resolution
// ends (game.MarkResolvingSpellReturnsToHandForEffect, Soulfire Grand Master's
// exit) — after the answer, since the question holds the resolution's tail. An
// empty library reveals nothing.
//
// No simplification.
func init() {
	const name = "Unexpected Results"
	effects.Register(effects.Spec{
		OracleID:     "048bdc4a-4d26-420c-88ed-23a073c30fff",
		Name:         name,
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			you, source := item.Controller, ctx.Source()
			if err := (effects.ShuffleLibrary{Player: you}).Apply(ctx); err != nil {
				return err
			}
			g := ctx.Game
			revealed := g.RevealTopOfLibraryForEffect(you, source, 1, name+" — reveal the top card of your library")
			if len(revealed) == 0 {
				return nil
			}
			c, ok := g.LookupCardForEffect(revealed[0])
			if !ok {
				return nil
			}
			if c.IsLand() {
				return effects.PutFromLibraryOntoBattlefield{
					Player:   you,
					Cards:    revealed,
					Match:    effects.Land(),
					Max:      1,
					Optional: true,
					Label:    name + " — you may put the land onto the battlefield and return " + name + " to its owner's hand",
					ThenTail: unexpectedResultsLandPut,
				}.Apply(ctx)
			}
			perm := effects.AnySpell.On(g.FreeCastFromLibraryPermission(you, source, name))
			perm.LapseOnPass = game.LapseStaysInHand
			if len(effects.FreeCastCandidates(g, &perm, revealed)) == 0 {
				return nil
			}
			g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
				Chooser:      you,
				Source:       source,
				Reason:       name,
				Cards:        revealed,
				Permission:   &perm,
				Keyword:      game.MayCastKeywordFreeCast,
				AcceptLabel:  "Cast it free",
				DeclineLabel: "Don't cast it",
			})
			return nil
		},
	})
}

// unexpectedResultsLandPut is "and return Unexpected Results to its owner's
// hand", behind the land's entry: nothing if it was declined or kept off the
// battlefield. An on-disk key: never renamed, never reused.
var unexpectedResultsLandPut = effects.PutFromLibraryTail("unexpected-results/return-to-hand",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		if len(res.Entered) == 0 {
			return nil
		}
		ctx.Game.MarkResolvingSpellReturnsToHandForEffect(res.Source)
		return nil
	})
