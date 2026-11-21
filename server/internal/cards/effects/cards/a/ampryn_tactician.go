package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ampryn Tactician — Creature — Human Soldier {2}{W}{W}, 3/3 (Magic Origins, #5549):
//
//	"When this creature enters, creatures you control get +1/+1 until end of turn."
//
// Inspiring Captain's entry: every creature you control as the trigger
// resolves gets +1/+1 until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ae19803-23a7-4724-a4cd-dc6f3c16fb24",
		Name:         "Ampryn Tactician",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ampryn Tactician — creatures you control get +1/+1 until end of turn",
				effects.Do(effects.BoostUntilEOT{
					Match: effects.And(effects.Creature(), effects.YouControl()),
					Power: 1, Toughness: 1,
					Label: "Ampryn Tactician — +1/+1",
				})),
		},
	})
}
