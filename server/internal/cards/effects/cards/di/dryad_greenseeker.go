package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dryad Greenseeker — Creature — Dryad {1}{G}, 1/3 (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"{T}: Look at the top card of your library. If it's a land card, you may
//	 reveal it and put it into your hand."
//
// Sarinth Steelseeker's look at the top card, land half only: a land is
// offered to the hand, revealed as it goes; anything else, or a declined land,
// stays on top. A card that left the library while the prompt waited is a new
// object and is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c43c1d7-1899-45c1-992b-443873890d78",
		Name:         "Dryad Greenseeker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Look at the top card of your library. If it's a land card, you may reveal it and put it into your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
				if len(looked) == 0 {
					return nil
				}
				c, ok := g.LookupCardForEffect(looked[0])
				if !ok || !c.IsLand() {
					return nil
				}
				return effects.MayChoice{
					Player:   ctx.Controller(),
					Question: "Dryad Greenseeker — reveal " + c.Name + " and put it into your hand?",
					YesLabel: "Put it into your hand",
					NoLabel:  "Don't",
					Then:     dryadGreenseekerTakeLand,
					Carry:    []uuid.UUID{looked[0]},
				}.Apply(ctx)
			},
		}},
	})
}

var dryadGreenseekerTakeLand = effects.MayYes("s82-old-b199/dryad-greenseeker-take-land",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		id := effects.CarryAt(r, 0)
		if !effects.BisonWhistleStillInLibrary(ctx.Game, id) {
			return nil
		}
		ctx.Game.RevealTopOfLibraryForEffect(r.Chooser, ctx.Source(), 1, "Dryad Greenseeker — reveal the top card")
		return ctx.Game.BounceToHandForEffect(id)
	})
