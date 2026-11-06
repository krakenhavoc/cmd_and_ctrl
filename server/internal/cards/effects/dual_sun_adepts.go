package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dual-Sun Adepts — Creature — Human Soldier {2}{W}, 2/2 (Edge of
// Eternities, #4461):
//
//	"Double strike
//	 {5}: Creatures you control get +1/+1 until end of turn."
//
// Double strike is a printed keyword. The pump is Dwarven Provisioner's:
// the set of creatures is fixed as the ability resolves (CR 611.2c), so a
// creature that enters later that turn does not get it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "155e01a1-5809-4be9-ad38-c11553a4a723",
		Name:            "Dual-Sun Adepts",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated: []ActivatedAbility{{
			Label:   "{5}: Creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Cost:    ManaCost("{5}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "Dual-Sun Adepts — creatures you control get +1/+1"}.Apply(NewContext(g, item))
			},
		}},
	})
}
