package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Castle Embereth — Land:
//
//	"This land enters tapped unless you control a Mountain.
//	 {T}: Add {R}.
//	 {1}{R}{R}, {T}: Creatures you control get +1/+0 until end of turn."
//
// The red Castle: the checkland clause with one land type, a plain mana
// ability, and a mass pump. The pump fixes the set of creatures as it
// resolves (CR 611.2c), so a creature that enters later this turn is not
// pumped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91fbb25b-8521-483f-88b0-77778d25f7fd",
		Name:         "Castle Embereth",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{b06EntersTappedUnlessLandType("mountain")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}{R}, {T}: Creatures you control get +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}{R}{R}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()), Power: 1,
					Label: "Castle Embereth — creatures you control get +1/+0",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
