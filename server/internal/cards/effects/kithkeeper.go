package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kithkeeper — Creature — Elemental {6}{W}, 3/3 (Lorwyn Eclipsed, #4458):
//
//	"Vivid — When this creature enters, create X 1/1 green and white
//	 Kithkin creature tokens, where X is the number of colors among
//	 permanents you control.
//	 Tap three untapped creatures you control: This creature gets +3/+0
//	 and gains flying until end of turn."
//
// Vivid is an ability word with no rules meaning (CR 207.2c). X is
// Glister Bairn's colour count (colorsAmongPermanentsYouControl), read as
// the trigger resolves (CR 608.2h); colourless counts nothing. The cost is
// Baylen's "tap N untapped creatures you control": it is not the {T}
// symbol, so a creature that arrived this turn pays it, Kithkeeper itself
// may be one of the three, and it does not target. The pump is locked to
// this permanent: a Kithkeeper that left does not pump a new object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c5766f9-befd-4426-9527-d50af0699ca1",
		Name:         "Kithkeeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kithkeeper — create X 1/1 green and white Kithkin creature tokens, X the colors among permanents you control (vivid)",
				Do(Instruction(func(ctx *Context) error {
					x := len(colorsAmongPermanentsYouControl(ctx.Game, ctx.Controller()))
					if x == 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 green and white Kithkin"), N: x}.Apply(ctx)
				}))),
		},
		Activated: []ActivatedAbility{{
			Label: "Tap three untapped creatures you control: This creature gets +3/+0 and gains flying until end of turn.",
			Purpose: game.Purpose{
				Answers: game.AnswerPump | game.AnswerCombatGrant,
				Pump:    &game.Pump{Power: 3, Keywords: []string{"flying"}},
			},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: TargetPermanent("three untapped creatures you control", Creature()),
				Label:  "three untapped creatures you control",
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := thisGetsUntilEndOfTurn(3, 0, "Kithkeeper — +3/+0 until end of turn")(g, item); err != nil {
					return err
				}
				return thisGainsKeywordUntilEndOfTurn("flying", "Kithkeeper — flying until end of turn")(g, item)
			},
		}},
	})
}
