package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a213c650-06cc-4ea8-86f8-78ccfe103b38",
		Name:         "Bosco, Just a Bear",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bosco, Just a Bear — create a Food token for each legendary creature you control",
				func(g *game.Game, item *game.StackItem) error {
					n := countControlled(g, item.Controller, func(c game.Card) bool { return c.IsCreature() && c.IsLegendary() })
					if n <= 0 {
						return nil
					}
					return CreateToken{Template: FoodToken(), N: n}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice a Food: Put two +1/+1 counters on Bosco. He gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{2}{G}"), game.AbilityCost{SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food"))}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := plusOneCountersOnThis(2)(g, item); err != nil {
					return err
				}
				return thisGainsKeywordUntilEndOfTurn("trample", "Bosco, Just a Bear — trample until end of turn")(g, item)
			},
		}},
	})
}
