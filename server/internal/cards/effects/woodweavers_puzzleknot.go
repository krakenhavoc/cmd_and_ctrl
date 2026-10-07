package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Woodweaver's Puzzleknot — Artifact {2}:
//
//	"When this artifact enters, you gain 3 life and get {E}{E}{E} (three
//	 energy counters).
//	 {2}{G}, Sacrifice this artifact: You gain 3 life and get {E}{E}{E}."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3334b218-01e3-47a3-8a28-5cc2decbedea",
		Name:         "Woodweaver's Puzzleknot",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Woodweaver's Puzzleknot — you gain 3 life and get {E}{E}{E}", gainLifeAndGetEnergy(3, 3)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice this artifact: You gain 3 life and get {E}{E}{E}.",
			Cost:    Plus(ManaCost("{2}{G}"), SacrificeThis()),
			Purpose: game.Purpose{Energy: 3},
			Effect:  gainLifeAndGetEnergy(3, 3),
		}},
	})
}
