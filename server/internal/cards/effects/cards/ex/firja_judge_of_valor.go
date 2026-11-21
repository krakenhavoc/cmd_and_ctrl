package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Firja, Judge of Valor — Legendary Creature — Angel Cleric {2}{W}{B}{B},
// 2/4:
//
//	"Flying, lifelink
//	 Whenever you cast your second spell each turn, look at the top three
//	 cards of your library. Put one of them into your hand and the rest
//	 into your graveyard."
//
// Commune with Evil's look-and-take over three cards (a mandatory pick,
// the rest put into the graveyard, not milled).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d5dcf8c-ecda-49ae-980c-c8c9016252d9",
		Name:            "Firja, Judge of Valor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Firja, Judge of Valor — look at the top three cards; put one into your hand and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.TakeFromLibraryToHand{
						Player: item.Controller,
						Cards:  g.LookAtTopOfLibraryForEffect(item.Controller, 3),
						Max:    1,
						Label:  "Firja, Judge of Valor — put one of them into your hand and the rest into your graveyard",
						Key:    firjaTake,
					}.Apply(ctx)
				}),
		},
	})
}

// firjaTake is the look's continuation: the rest into the graveyard. An
// on-disk key.
var firjaTake = effects.RegisterTake("choose-cards/firja-judge-of-valor-take",
	effects.TakeDef{Then: func(g *game.Game, res effects.TakeFromLibraryResult) error {
		return effects.TakeRestIntoGraveyard(g, res)
	}})
