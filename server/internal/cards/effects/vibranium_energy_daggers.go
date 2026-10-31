package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vibranium Energy Daggers — {1} Artifact — Equipment (Marvel Super
// Heroes, #4455):
//
//	"Indestructible
//	 Equipped creature gets +2/+2.
//	 Equip {3}"
//
// Darksteel Plate's shape: indestructible is a printed keyword of the
// Equipment itself, the pump is the attachment static.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be0b86fc-8450-41d4-9836-6de11cecb857",
		Name:            "Vibranium Energy Daggers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Static:          []game.StaticAbility{PumpAttached(2, 2)},
		Activated:       []ActivatedAbility{EquipAbility("{3}")},
	})
}
