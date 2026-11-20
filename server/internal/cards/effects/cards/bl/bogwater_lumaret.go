package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bogwater Lumaret — Creature — Spirit Frog {B}{G}, 2/2, Secrets of Strixhaven
// (#4456):
//
//	"Whenever this creature or another creature you control enters, you gain
//	 1 life."
//
// One trigger on every creature entering under your control, itself included
// (CreatureEnteredUnderYourControl reads the entering object's layered type,
// so an animated artifact counts and an opponent's creature does not). It
// fires once per creature, so a mass token creation gains once per token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0109432d-5a2a-456f-ad39-b75cb4c73420",
		Name:         "Bogwater Lumaret",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
				"Bogwater Lumaret — you gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
