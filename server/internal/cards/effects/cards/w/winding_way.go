package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Winding Way — Sorcery {1}{G} (Outlaws of Thunder Junction Commander, #5368):
//
//	"Choose creature or land. Reveal the top four cards of your library. Put
//	 all cards of the chosen type revealed this way into your hand and the rest
//	 into your graveyard."
//
// The type is asked first (a two-option pick), then the top four are revealed,
// every card of the chosen type goes to the hand and the rest are put into the
// graveyard (not milled). Mulch's body with the pick in front.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e48e76c-8756-4378-a722-2e2929feb74f",
		Name:         "Winding Way",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ctx.Game.QueueOptionPickForEffect(game.OptionPickPrompt{
				Chooser:  item.Controller,
				Source:   ctx.Source(),
				Question: "Winding Way — choose creature or land",
				Options:  []game.ChoiceOption{{Label: "Creature"}, {Label: "Land"}},
				ThenKey:  windingWayChosen,
			})
			return nil
		},
	})
}

// windingWayChosen reveals the top four and takes every card of the chosen
// type. An on-disk key: never renamed, never reused.
var windingWayChosen = game.RegisterOptionPickThen("option-pick/winding-way-type", func(g *game.Game, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	match := effects.Land()
	if r.Option.Label == "Creature" {
		match = effects.Creature()
	}
	return effects.Cmdr2B15RevealTopTakeAllMatching(g, r.Chooser, r.Source, 4,
		"Winding Way — reveal the top four cards, "+r.Option.Label+" cards to your hand and the rest to your graveyard", match)
})
