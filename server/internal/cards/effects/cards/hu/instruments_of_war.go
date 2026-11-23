package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Instruments of War — Artifact {4} (Jumpstart 2022, #5948):
//
//	"Flash
//	 As this artifact enters, choose a creature type.
//	 Creatures you control of the chosen type get +1/+1."
//
// Adaptive Automaton's choice and anthem, without the "other" restriction:
// the Instruments is not a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8be46af9-1b70-4da9-b16d-869c3831dc99",
		Name:            "Instruments of War",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		AsEnters:        effects.ChooseCreatureTypeAsEnters("Instruments of War"),
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Chosen: true, YoursOnly: true}, 1, 1),
		},
	})
}
