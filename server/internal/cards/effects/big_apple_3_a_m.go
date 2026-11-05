package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Big Apple, 3 a.m. — Land (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"This land enters tapped. As it enters, choose a color.
//	 {T}: Add one mana of the chosen color.
//	 {5}, {T}: Create a 1/1 black Rat creature token for each opponent
//	 you have."
//
// The first two lines are Uncharted Haven's (chosen_color_lands.go): a
// tapped entry, an as-enters colour prompt and a mana ability that reads
// the stored colour. The Rats are counted as the ability resolves, one
// for each opponent still in the game.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd01ef1f-f6be-498f-82e0-dc04833e685f",
		Name:         "Big Apple, 3 a.m.",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, "Big Apple, 3 a.m."),
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			ProducedFunc: ProducedChosenColor(),
			Label:        "Add one mana of the chosen color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Create a 1/1 black Rat creature token for each opponent you have.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Effect: Do(CreateTokensForEach{
				Template: TokenCard("1/1 black Rat"),
				Name:     "each opponent you have",
				Count:    func(ctx *Context) int { return len(ctx.Opponents()) },
			}),
		}},
	})
}
