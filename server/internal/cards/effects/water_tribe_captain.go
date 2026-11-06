package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Water Tribe Captain — Creature — Human Soldier Ally {2}{W}, 3/3
// (Avatar: The Last Airbender, #4459):
//
//	"{5}: Creatures you control get +1/+1 until end of turn."
//
// The affected set is locked as the ability resolves (CR 611.2c): a
// creature that enters afterwards does not get the bonus.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c6c4921-c6cb-4b4b-ac9d-10ccc89a5d1d",
		Name:         "Water Tribe Captain",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}: Creatures you control get +1/+1 until end of turn.",
			Cost:    ManaCost("{5}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{
					Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "Water Tribe Captain — creatures you control get +1/+1",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
