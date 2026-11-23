package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bosco, Just a Bear — Legendary Creature — Bear {4}{G}, 4/4 (Avatar:
// The Last Airbender Eternal, #5075):
//
//	"When Bosco enters, create a Food token for each legendary creature
//	 you control.
//	 {2}{G}, Sacrifice a Food: Put two +1/+1 counters on Bosco. He gains
//	 trample until end of turn."
//
// The count is read as the enters trigger resolves, over the layered
// battlefield, so Bosco counts himself while he is there (the ruling)
// and a legendary creature that left in response does not.
//
// "Sacrifice a Food" is Gyome, Master Chef's cost: any Food artifact
// you control, a token or not, by its effective subtype. The body is
// Rough Rhino Cavalry's: two counters and trample on Bosco only while he
// is still the permanent that activated it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a213c650-06cc-4ea8-86f8-78ccfe103b38",
		Name:         "Bosco, Just a Bear",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bosco, Just a Bear — create a Food token for each legendary creature you control",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() && c.IsLegendary() })
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.FoodToken(), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice a Food: Put two +1/+1 counters on Bosco. He gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Food", effects.HasSubtype("Food"))}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := effects.PlusOneCountersOnThis(2)(g, item); err != nil {
					return err
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("trample", "Bosco, Just a Bear — trample until end of turn")(g, item)
			},
		}},
	})
}
