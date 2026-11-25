package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wishful Merfolk — Creature — Merfolk {1}{U}, 3/2 (Jumpstart, #5510):
//
//	"Defender
//	 {1}{U}: This creature loses defender and becomes a Human until end of
//	 turn."
//
// One layer-6 effect removes the keyword and sets the creature types to
// Human (CR 205.1a: "becomes a Human" replaces the Merfolk type), the way
// Kellan, Planar Trailblazer changes its types. It lasts until end of turn
// and the creature stays the same object.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "19ff1424-6505-4702-86ad-2d7518b584f6",
		Name:            "Wishful Merfolk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: This creature loses defender and becomes a Human until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				return effects.UntilEndOfTurn(effects.NewContext(g, item), item.SourceCardID, nil, "Wishful Merfolk — loses defender and becomes a Human",
					game.RemoveKeywordsMod("defender"),
					game.SetCreatureTypesMod(g.TextReaderForEffect(item).TextSubtypes("Human")...))
			},
		}},
	})
}
