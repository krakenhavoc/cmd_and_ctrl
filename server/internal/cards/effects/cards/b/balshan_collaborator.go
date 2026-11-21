package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Balshan Collaborator — Creature — Bird Soldier {3}{U}, 2/2 (Torment,
// #5612):
//
//	"Flying
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Flying is a printed keyword (CR 702.9b). The pump is the shared
// selfPumpRow Nantuko Shade uses: no tap, so any number of activations,
// each ending at cleanup (CR 514.2). The {B} is an off-colour cost, paid
// from black mana like any other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7cdbc115-cb15-4ada-8229-d40f0b22dd27",
		Name:            "Balshan Collaborator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.SelfPumpRow("Balshan Collaborator", "{B}", 1, 1)},
	})
}
