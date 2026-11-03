package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0109432d-5a2a-456f-ad39-b75cb4c73420",
		Name:         "Bogwater Lumaret",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl,
				"Bogwater Lumaret — you gain 1 life", Do(GainLife{Amount: 1})),
		},
	})
}
