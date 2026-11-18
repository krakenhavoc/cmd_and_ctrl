package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dueling Rapier — Artifact — Equipment {R} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Flash
//	 When this Equipment enters, attach it to target creature you control.
//	 Equipped creature gets +2/+0.
//	 Equip {4}"
//
// Paladin's Shield's flash and enter-attach shape (afrB06EnterAttach).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7ceec1fd-4af5-40d4-9d17-424e21889d64",
		Name:            "Dueling Rapier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{PumpAttached(2, 0)},
		Triggered:       []game.TriggeredAbility{afrB06EnterAttach("Dueling Rapier")},
		Activated:       []ActivatedAbility{EquipAbility("{4}")},
	})
}
