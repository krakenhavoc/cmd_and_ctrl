package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marshland Bloodcaster — Creature — Vampire Warlock {4}{B}, 3/5 (Outlaws of
// Thunder Junction Commander):
//
//	"Flying
//	 {1}{B}, {T}: Rather than pay the mana cost of the next spell you cast
//	 this turn, you may pay life equal to that spell's mana value."
//
// Apex Observatory's next-spell promise with a life price
// (game.NextSpellPromise.PayLifeEqualToManaValue): the next spell cast this
// turn is offered one more alternative cost (CR 118.9), paid in life
// (CR 119.4) — {X} counts as 0 (CR 107.3b), and a spell with no mana cost
// costs none. The promise is spent by the next spell whichever cost it is cast
// for, and is only an offer: the spell may still be paid for with mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ecf35b1-0bfd-4c60-a791-5a3b918dd630",
		Name:            "Marshland Bloodcaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}: Rather than pay the mana cost of the next spell you cast this turn, you may pay life equal to that spell's mana value.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GrantNextSpellPromise{From: "Marshland Bloodcaster", Promise: game.NextSpellPromise{
					PayLifeEqualToManaValue: true,
					Text:                    "Rather than pay the mana cost of the next spell you cast this turn, you may pay life equal to that spell's mana value.",
				}}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
