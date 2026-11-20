package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ampryn Tactician — Creature — Human Soldier {2}{W}{W}, 3/3 (Magic Origins, #5549):
//
//	"When this creature enters, creatures you control get +1/+1 until end of turn."
//
// Inspiring Captain's entry: every creature you control as the trigger
// resolves gets +1/+1 until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ae19803-23a7-4724-a4cd-dc6f3c16fb24",
		Name:         "Ampryn Tactician",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ampryn Tactician — creatures you control get +1/+1 until end of turn",
				Do(BoostUntilEOT{
					Match: And(Creature(), YouControl()),
					Power: 1, Toughness: 1,
					Label: "Ampryn Tactician — +1/+1",
				})),
		},
	})
}
