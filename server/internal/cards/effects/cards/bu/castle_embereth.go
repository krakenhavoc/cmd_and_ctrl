package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "91fbb25b-8521-483f-88b0-77778d25f7fd",
		Name:         "Castle Embereth",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B06EntersTappedUnlessLandType("mountain")},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{R}{R}, {T}: Creatures you control get +1/+0 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1}, Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.ManaCost("{1}{R}{R}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
					Label: "Castle Embereth — creatures you control get +1/+0",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
