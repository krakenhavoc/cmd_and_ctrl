package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "7ceec1fd-4af5-40d4-9d17-424e21889d64",
		Name:            "Dueling Rapier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Static:          []game.StaticAbility{effects.PumpAttached(2, 0)},
		Triggered:       []game.TriggeredAbility{effects.AfrB06EnterAttach("Dueling Rapier")},
		Activated:       []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
