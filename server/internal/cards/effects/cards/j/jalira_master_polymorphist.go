package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jalira, Master Polymorphist — Legendary Creature — Human Wizard {3}{U}, 2/2
// (Magic 2015, #5556):
//
//	"{2}{U}, {T}, Sacrifice another creature: Reveal cards from the top of
//	 your library until you reveal a nonlegendary creature card. Put that
//	 card onto the battlefield and the rest on the bottom of your library in
//	 a random order."
//
// Bag of Tricks' reveal-until body (RevealUntilThenPutOntoBattlefield) with
// the sacrifice of another creature as a cost, the tap and the mana. A
// library with no nonlegendary creature card reveals itself whole and puts it
// all on the bottom, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62ad15f1-ee51-4991-8e52-06a3cbd8f395",
		Name:         "Jalira, Master Polymorphist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}, Sacrifice another creature: Reveal cards from the top of your library until you reveal a nonlegendary creature card. Put that card onto the battlefield and the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerProtect | game.AnswerSurvive},
			Cost: effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost(),
				effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RevealUntilThenPutOntoBattlefield{
					Match:  func(c game.Card) bool { return c.IsCreature() && !c.HasSupertype("legendary") },
					Reason: "Jalira, Master Polymorphist — revealed until a nonlegendary creature card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
