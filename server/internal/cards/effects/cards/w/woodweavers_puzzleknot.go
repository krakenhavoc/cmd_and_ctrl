package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "3334b218-01e3-47a3-8a28-5cc2decbedea",
		Name:         "Woodweaver's Puzzleknot",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Woodweaver's Puzzleknot — you gain 3 life and get {E}{E}{E}", effects.GainLifeAndGetEnergy(3, 3)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}, Sacrifice this artifact: You gain 3 life and get {E}{E}{E}.",
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.SacrificeThis()),
			Purpose: game.Purpose{Answers: game.AnswerValue, Energy: 3},
			Effect:  effects.GainLifeAndGetEnergy(3, 3),
		}},
	})
}
